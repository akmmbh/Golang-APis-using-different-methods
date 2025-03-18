package simplersa

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
)

//public key is part of RSA key pair
type PublicKey struct{
	N *big.Int
	E *big.Int
}
//private key also is part 
//could include the prime factor of N and other data to make decryption
type PrivateKey struct{
	N *big.Int
	D *big.Int 
}
//generate a public and private key pair for rsa
func GenerateKeys(bitlen int)(*PublicKey,*PrivateKey,error){
	numRetries:=0;
	for{
		numRetries++
		if numRetries ==10{
			panic("retrying too many time ,something is wrong")

		}
		//we need result pq with b bits so we generate p and q with b/2 bites
		//each if the otp bit p and q are set the result will have b bites
		//other we will retry and rand.Prime should return prime with their top
		// bit set so in practice there will be bo retries
		p,err:= rand.Prime(rand.Reader,bitlen/2)
		if err!=nil{
			return nil,nil,err
		}
		q,err := rand.Prime(rand.Reader,bitlen/2)
		if err!=nil{
			return nil,nil,err
		}
		n:= new(big.Int).Set(p)
		n.Mul(n,q)
		if n.BitLen()!= bitlen{
			continue;
		}
		p.Sub(p,big.NewInt(1))
		q.Sub(q,big.NewInt(1))
		totient:=new(big.Int).Set(p)
		totient.Mul(totient,q)
		e:= big.NewInt(65537)
		//calculate the modular multiplicatigve inverse of e such taht 
		//de =1 (mod totient)
		//if gcd(e,totient)=1 then e is guranteed to have unique inverse but
		//since p-1 and q-1 could theoritically have e as a factor this may fail 
		//once in a while (likely to a execeeding rase)
		d:= new(big.Int).ModInverse(e,totient)
		if d ==nil{
			continue
		}
		pub := &PublicKey{N:n ,E:e}
		priv := &PrivateKey{N:n,D:d}
		return pub ,priv,nil
	}
}
func encrypt(pub *PublicKey,m *big.Int)*big.Int{
	c:=new(big.Int)
	c.Exp(m,pub.E,pub.N)
	return c
}
//decrypt perform decrypption of the cipher c using a private key and the
//decrytped message
func decrypt(priv *PrivateKey,c *big.Int)*big.Int{
	m:=new(big.Int)
	m.Exp(c,priv.D,priv.N)
	return m
}
//encrypt rsa encrypt the message m using pubolic key and return the encrypt
//bits the length of must be <= size_in_bytes(pub.N)-11,

//otherwise error is returned 
func EncryptRSA(pub *PublicKey, m []byte)([]byte,error){
	//comnpute lenght of th ekey in bytes rounding up
	keyLen:=(pub.N.BitLen()+7)/8
	if len(m)>keyLen-11{
		return nil,fmt.Errorf("len(m)=%v,toolong",len(m))

	}
	//following RFC 2313 using block type 02 as recommended for encryption
	//EB = 00 ||02||PS ||00||D
	//PS = padding D = Data
	psLen:= keyLen-len(m)-3
	eb:=make([]byte,keyLen)
	eb[0]=0x00
	eb[1]=0x02
	for i:=2; i<2+psLen;{
    _,err:=rand.Read(eb[i:i+1])
	if err!=nil{
		return nil,err
	}
	if eb[i]!=0x00{
		i++
	}
	}
	eb[2+psLen]=0x00
	//copy message m into rest of encryption block
	copy(eb[3+psLen:],m)
	//now the encryption block is compelte we cantake mbyte big.Int
	//Rsa encrypt it with public key
	mnum:=new(big.Int).SetBytes(eb)
	c:=encrypt(pub,mnum)
	//the result is a big Int which we want to convert to a byte of slice
	//length key len it s hightly likely that the size of c in bytes is keylen
	//but in rare cases we may need to pad it ont hte left with zeros 
	padLen := keyLen - len(c.Bytes())
	for i := 0; i < padLen; i++ {
		eb[i] = 0x00
	}
	copy(eb[padLen:], c.Bytes())
	return eb, nil
}
//Decrypt decrypts the message c using private key priv and return the
//decrypted bytes based on block 02 from the PKCS #1 
//it expects the length in bytes of the private key modulo to be len(eb)
//important this is simple implementation designes to be resileent to 
//timing attacks
func DecryptRSA(priv *PrivateKey, c  []byte)([]byte,error){
	keyLen:=(priv.N.BitLen()+7)/8
	if len(c)!=keyLen{
		return nil,fmt.Errorf("len(c)=%v,want keyLen=%v",len(c),keyLen)

	}
	//convert cinto bit.Int and decrypt it using the private key
	cnum:= new(big.Int).SetBytes(c)
	mnum:=decrypt(priv,cnum)
	m := make([]byte, keyLen)
	copy(m[keyLen-len(mnum.Bytes()):], mnum.Bytes())

	// Expect proper block 02 beginning.
	if m[0] != 0x00 {
		return nil, fmt.Errorf("m[0]=%v, want 0x00", m[0])
	}
	if m[1] != 0x02 {
		return nil, fmt.Errorf("m[1]=%v, want 0x02", m[1])
	}

	// Skip over random padding until a 0x00 byte is reached. +2 adjusts the index
	// back to the full slice.
	endPad := bytes.IndexByte(m[2:], 0x00) + 2
	if endPad < 2 {
		return nil, fmt.Errorf("end of padding not found")
	}

	return m[endPad+1:], nil
}
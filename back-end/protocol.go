package backend

import (
	"bytes"
	_ "bytes"
	"crypto/ecdsa"
	_ "crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	_ "errors"
	"fmt"
	_ "fmt"
	"log"
	_ "log"
	"math/big"
	ec "swag/ec"
	"github.com/binance-chain/tss-lib/crypto"
)


type Reviewer struct {
	keys                *ecdsa.PrivateKey
	biddedPaperMap      map[int][]byte
	paperMap            map[int][]byte
	signatureMap        map[int][]byte
	paperCommittedValue *Paper
}

type Submitter struct {
	keys                    *ecdsa.PrivateKey
	userID                  string
	submitterCommittedValue *CommitStruct //commitstruct
	paperCommittedValue     *Paper
	receiver                *Receiver
	encrypted               []byte
	signatureMap            map[int][]byte
}

type CommitStruct struct {
	CommittedValue *ecdsa.PublicKey
	r              *big.Int
	val            *big.Int
}

type PC struct {
	keys         *ecdsa.PrivateKey
	signatureMap map[int][]byte
}

type Paper struct {
	Id                  int
	CommittedValue      *CommitStruct
	Selected            bool
	ReviewSignatureByPC []byte
}



var (
	tree = NewTree(DefaultMinItems)
	pc = PC{
		newKeys(),
		nil,
	}
	paperList     []Paper
	schnorrProofs []SchnorrProof
)

type SubmitStruct struct {
	Msg       []byte
	Rr        []byte
	Rs        []byte
	SharedKey []byte
}

type Receiver struct {
	keys       *ecdsa.PrivateKey
	commitment *ecdsa.PublicKey
}

func NewReceiver(key *ecdsa.PrivateKey) *Receiver {
	return &Receiver{
		keys: key,
	}
}

func SetCommitment(r *Receiver, comm *ecdsa.PublicKey) {
	r.commitment = comm
}

func GetTrapdoor(r *Receiver) *big.Int {
	return r.keys.D
}

// It returns values x and r (commitment was c = g^x * g^r).
func (s *Submitter) GetDecommitMsg() (*big.Int, *big.Int) {
	val := s.submitterCommittedValue.val
	r := s.submitterCommittedValue.r

	return val, r
}

func (s *Submitter) GetDecommitMsgPaper() (*big.Int, *big.Int) {
	val := s.paperCommittedValue.CommittedValue.val
	r := s.paperCommittedValue.CommittedValue.r

	return val, r
}

// When receiver receives a decommitment, CheckDecommitment verifies it against the stored value
// (stored by SetCommitment).
func (r *Receiver) CheckDecommitment(R, val *big.Int) bool {
	a := ec.ExpBaseG(r.keys, val)             // g^x
	b := ec.Exp(r.keys, &r.keys.PublicKey, R) // h^r
	c := ec.Mul(r.keys, a, b)                 // g^x * h^r

	return Equals(c, r.commitment)
}

func generateSharedSecret(pc *PC, submitter *Submitter, reviewer *Reviewer) string {
	publicPC := pc.keys.PublicKey
	var sharedHash [32]byte
	if reviewer == nil {
		privateS := submitter.keys
		shared, _ := publicPC.Curve.ScalarMult(publicPC.X, publicPC.Y, privateS.D.Bytes())
		sharedHash = sha256.Sum256(shared.Bytes())
	} else {
		privateR := reviewer.keys
		shared, _ := publicPC.Curve.ScalarMult(publicPC.X, publicPC.Y, privateR.D.Bytes())
		sharedHash = sha256.Sum256(shared.Bytes())
	}
	
	return string(sharedHash[:])
}

func newKeys() *ecdsa.PrivateKey {
	a, _ := ecdsa.GenerateKey(curve, rand.Reader)
	return a
}

func Submit(s *Submitter, p *Paper) *Submitter {
	rr := GetRandomInt(s.keys.D)
	rs := GetRandomInt(s.keys.D)
	ri := GetRandomInt(s.keys.D)
	
	log.Printf("\n, %s", "Generate rr from s.keys.D and storing in btree/log")
	tree.Put("Rr", rr)
	log.Printf("\n, %s", "Generate rs from s.keys.D and storing in btree/log")
	tree.Put("Rs", rs)
	log.Printf("\n, %s", "Generate ri from s.keys.D and storing in btree/log")
	tree.Put("Ri", ri)
	
	log.Println(rr) // shared between all parties
	log.Println(rs) // shared between S and PC
	log.Println(ri) // step 2

	sharedPCS := generateSharedSecret(&pc, s, nil)

	hashedPublicK := sha256.Sum256(EncodeToBytes(pc.keys.PublicKey.X))
	encryptedSharedPCS := Encrypt([]byte(sharedPCS), string(hashedPublicK[:]))

	msg := SubmitStruct{
		Encrypt(EncodeToBytes(p), sharedPCS),
		Encrypt(EncodeToBytes(rr), sharedPCS),
		Encrypt(EncodeToBytes(rs), sharedPCS),
		encryptedSharedPCS,
	}
	
	tree.Put("msg", msg)
	log.Println("Encrypted paper and random values logged")

	s.encrypted = Encrypt(EncodeToBytes(msg), s.keys.D.String()) //encrypted paper and random numbers

	//submitter identity commit
	s.GetCommitMessage(ri)

	//paper identity commit
	s.GetCommitMessagePaper(rs)

	hashedMsgSubmit, _ := GetMessageHash([]byte(fmt.Sprintf("%v", s.submitterCommittedValue.CommittedValue)))
	hashedMsgPaper, _ := GetMessageHash([]byte(fmt.Sprintf("%v", s.paperCommittedValue.CommittedValue.CommittedValue)))

	signatureSubmit, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedMsgSubmit)
	putNextSignatureInMapSubmitter(s, signatureSubmit)

	signaturePaper, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedMsgPaper)
	putNextSignatureInMapSubmitter(s, signaturePaper)

	log.Printf("\n %s %s", "Ks is revealed to all parties", s.keys.PublicKey) //KS is logged/revealed to all parties??? or is it

	hashedPaperPC, _ := GetMessageHash([]byte(fmt.Sprintf("%v", s.paperCommittedValue.CommittedValue.CommittedValue)))
	signaturePaperPC, _ := ecdsa.SignASN1(rand.Reader, pc.keys, hashedPaperPC)
	putNextSignatureInMapPC(&pc, signaturePaperPC)                            //signal next fase
	log.Println("PC signed a paper (submission) " + string(signaturePaperPC)) //PC signed paper commit to indicate the PC will continue the process of getting the paper reviewed

	paperList = append(paperList, *p)

	return s
}

func putNextSignatureInMapSubmitter(s *Submitter, slice []byte) { //not sure if works, test needed.
	for k, v := range s.signatureMap {
		if v == nil {
			s.signatureMap[k] = slice
		}
	}
}

func putNextSignatureInMapPC(p *PC, slice []byte) {
	for k, v := range p.signatureMap {
		if v == nil {
			pc.signatureMap[k] = slice
		}
	}
}

func putNextSignatureInMapReviewer(r *Reviewer, slice []byte) {
	for k, v := range r.signatureMap {
		if v == nil {
			r.signatureMap[k] = slice
		}
	}
}

func putNextPaperInBidMapReviewer(r *Reviewer, slice []byte) {
	for k, v := range r.biddedPaperMap {
		if v == nil {
			r.biddedPaperMap[k] = slice
		}
	}
}

func GetMessageHash(xd []byte) ([]byte, error) {
	md := sha256.New()
	return md.Sum(xd), nil
}

//step 4
func assignPapers(pc *PC, reviewerSlice []Reviewer, paperSlice []Paper) {
	for r := range reviewerSlice {
		Kpcr := generateSharedSecret(pc, nil, &reviewerSlice[r])
		for p := range paperSlice {

			hashedPaper, _ := GetMessageHash(EncodeToBytes(paperSlice[p]))
			pcSignature, _ := ecdsa.SignASN1(rand.Reader, pc.keys, hashedPaper)
			putNextSignatureInMapPC(pc, pcSignature)
			toBeEncrypted := EncodeToBytes(paperSlice[p])
			encrypted := Encrypt(toBeEncrypted, Kpcr)
			reviewerSlice[r].paperMap[p] = encrypted
		}
	}
}

func getPaperList(pc *PC, reviewer *Reviewer) []Paper {

	pMap := reviewer.paperMap
	Kpcr := generateSharedSecret(pc, nil, reviewer)
	pList := []Paper{}
	for _, v := range pMap {
		decrypted := Decrypt(v, Kpcr)
		p := DecodeToPaper(decrypted)
		pList = append(pList, p)
	}
	return pList
}

func makeBid(r *Reviewer, pap *Paper) {
	pList := getPaperList(&pc, r)

	for _, p := range pList {
		if p.Id == pap.Id {
			p.Selected = true
		}
	}
}

//step 5
func setEncBidList(r *Reviewer) { //set encrypted bid list

	pList := getPaperList(&pc, r)
	Kpcr := generateSharedSecret(&pc, nil, r)
	tmpPaperList := []Paper{}
	for _, p := range pList {
		if p.Selected == true {
			tmpPaperList = append(tmpPaperList, p)
			putNextPaperInBidMapReviewer(r, Encrypt(EncodeToBytes(p), Kpcr))
		}
	}

	hashedBiddedPaperList, _ := GetMessageHash(EncodeToBytes(r.biddedPaperMap)) //changed from tmpPaperList to r.bidedPaperMap
	rSignature, _ := ecdsa.SignASN1(rand.Reader, r.keys, hashedBiddedPaperList)
	putNextSignatureInMapReviewer(r, rSignature)

	//r.biddedPaperMap = Encrypt(EncodeToBytes(tmpPaperList), Kpcr)
}

func matchPaper(reviewerSlice []Reviewer) { //step 6 (some of it)

	pList := getPaperList(&pc, &reviewerSlice[0])

	for _, rev := range reviewerSlice {
		kcpr := generateSharedSecret(&pc, nil, &rev)
		for i := range rev.biddedPaperMap {
			decrypted := Decrypt(rev.biddedPaperMap[i], kcpr)
			paper := DecodeToPaper(decrypted)
			if paper.Selected {
				for i, p := range pList {
					if paper.Id == p.Id {
						rev.paperCommittedValue = &paper      //assigning paper
						pList[i] = Paper{-1, nil, false, nil} //removing paper from generic map
						break
					}
				}
			}
		}
	}
	// Case for if reviewer weren't assigned paper because
	// other reviewer might have gotten the paper beforehand
	for _, rev := range reviewerSlice {
		if rev.paperCommittedValue == &(Paper{}) {
			for i, v := range pList {
				if pList[i].Id != -1 { // checking for removed papers
					//if this statement is true we have a normal paper
					rev.paperCommittedValue = &v
					pList[i] = Paper{-1, nil, false, nil} //removing paper from generic map
					break
				}
			}
		} else {
			break
		}
	}
}

func finalMatching(reviewers []Reviewer, submitters []Submitter) {
	for _, r := range reviewers {
		r.GetCommitMessageReviewPaper(GetRandomInt(r.keys.D))
		nonce := GetRandomInt(r.keys.D)
		log.Printf("\n, %s, %s", "Nonce, step 6:", nonce)
		hash, _ := GetMessageHash(EncodeToBytes(r.keys.PublicKey))
		signature, _ := ecdsa.SignASN1(rand.Reader, pc.keys, hash)
		putNextSignatureInMapPC(&pc, signature)

		for _, s := range submitters {
			paperCommitSubmitter := s.paperCommittedValue.CommittedValue.CommittedValue
			paperCommitReviewer := r.paperCommittedValue.CommittedValue.CommittedValue
			if paperCommitSubmitter == paperCommitReviewer {
				schnorrProofs = append(schnorrProofs, *CreateProof(s.keys, r.keys)) //NOT CORRECT, WAIT FOR ANSWER FROM SUPERVISOR

			}
		}
	}
}

func EcdsaToECPoint(pk *ecdsa.PublicKey) (*crypto.ECPoint, error) {
	return crypto.NewECPoint(pk.Curve, pk.X, pk.Y)
}

// func ECPointToEcdsa(ec *crypto.ECPoint) (*ecdsa.PublicKey){
// 	return &ecdsa.PublicKey{ec.Curve(), ec[0], ec.coords[1]}
// }

func (s *Submitter) GetCommitMessage(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(s.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := GetRandomInt(s.keys.D)

	s.submitterCommittedValue.r = r     //hiding factor?
	s.submitterCommittedValue.val = val //den value (random) vi comitter ting til
	x1 := ec.ExpBaseG(s.keys, val)
	x2 := ec.Exp(s.keys, &s.keys.PublicKey, r)
	comm := ec.Mul(s.keys, x1, x2)
	s.submitterCommittedValue.CommittedValue = comm

	return comm, nil
} //C(P, r)  C(S, r)

func (rev *Reviewer) GetCommitMessageReviewPaper(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(rev.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := GetRandomInt(rev.keys.D)

	rev.paperCommittedValue.CommittedValue.r = r

	rev.paperCommittedValue.CommittedValue.val = val

	x1 := ec.ExpBaseG(rev.keys, val)
	x2 := ec.Exp(rev.keys, &rev.keys.PublicKey, r)
	comm := ec.Mul(rev.keys, x1, x2)
	rev.paperCommittedValue.CommittedValue.CommittedValue = comm
	fmt.Printf("\n %s, %s, %s", "R & Val (Reviewer): ", r, val)
	fmt.Printf("\n %s, %s", "comm (Reviewer)", comm)

	return comm, nil
} //C(P, r)  C(S, r)

func (s *Submitter) GetCommitMessagePaper(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(s.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := GetRandomInt(s.keys.D) //check up on this

	s.paperCommittedValue.CommittedValue.r = r

	s.paperCommittedValue.CommittedValue.val = val

	x1 := ec.ExpBaseG(s.keys, val)
	x2 := ec.Exp(s.keys, &s.keys.PublicKey, r)
	comm := ec.Mul(s.keys, x1, x2)
	s.paperCommittedValue.CommittedValue.CommittedValue = comm
	fmt.Printf("\n %s, %s, %s", "R & Val: (Submitter)", r, val)
	fmt.Printf("\n %s, %s", "comm (Submitter)", comm)
	return comm, nil
}

//verify
func (s *Submitter) VerifyTrapdoorSubmitter(trapdoor *big.Int) bool {
	h := ec.ExpBaseG(s.keys, trapdoor)
	return Equals(h, &s.keys.PublicKey)
	//Equals(key, &s.keys.PublicKey)
}

/*
//verify
func (s *Submitter) VerifyTrapdoorPaper(trapdoor *big.Int) bool {
	h:= ec.ExpBaseG(s.keys, s.keys.D)
	return Equals(h, &s.Pa)

	hx, hy := p.CommittedValue.Curve.ScalarBaseMult(trapdoor.Bytes())
	key := &ecdsa.PublicKey{p.CommittedValue.Curve, hx, hy}
	return key.Equal(p.CommittedValue)
	//Equals(key, &s.keys.PublicKey)

}*/

func Equals(e *ecdsa.PublicKey, b *ecdsa.PublicKey) bool {
	return e.X.Cmp(b.X) == 0 && e.Y.Cmp(b.Y) == 0
}

func EncodeToBytes(p interface{}) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("uncompressed size (bytes): ", len(buf.Bytes()))
	return buf.Bytes()
}

func DecodeToPaper(s []byte) Paper {

	p := Paper{}
	dec := gob.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&p)
	if err != nil {
		log.Fatal(err)
	}
	return p
}

func GetRandomInt(max *big.Int) *big.Int {
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Fatal(err)
	}
	return n
}
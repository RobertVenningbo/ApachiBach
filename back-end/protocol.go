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
	"strconv"
	ec "swag/ec"
	"github.com/binance-chain/tss-lib/crypto"
)


type Reviewer struct {
	userID				string
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
	
	sharedKpcs := generateSharedSecret(&pc, s, nil)  //Shared secret key between Submitter and PC (Kpcs)

	hashedPublicK := sha256.Sum256(EncodeToBytes(pc.keys.PublicKey.X)) //PC's hashed public key
	encryptedSharedKpcs := Encrypt([]byte(sharedKpcs), string(hashedPublicK[:])) //Encrypted Kpcs with PC's public key

	EncryptedPaperAndRandomness := SubmitStruct{ //Encrypted Paper and Random numbers
		Encrypt(EncodeToBytes(p), sharedKpcs),
		Encrypt(EncodeToBytes(rr), sharedKpcs),
		Encrypt(EncodeToBytes(rs), sharedKpcs),
		encryptedSharedKpcs,
	}
	LoggedMessage := fmt.Sprintf("%#v", EncryptedPaperAndRandomness)
	tree.Put(LoggedMessage, EncryptedPaperAndRandomness)
	log.Println(LoggedMessage + " - Encrypted Paper and Random Numbers logged")
	

	s.encrypted = Encrypt(EncodeToBytes(EncryptedPaperAndRandomness), s.keys.D.String()) //TODO: Do we need  this if we log it above??
	
	SubmissionSignature, _ := ecdsa.SignASN1(rand.Reader, s.keys, s.encrypted) //Entire message signed by submission private key
	SubmitterAsString := fmt.Sprintf("%#v", s)
	tree.Put(SubmitterAsString + s.userID + "SubmissionSignature", SubmissionSignature)
	log.Println(SubmitterAsString + s.userID + " " + string(SubmissionSignature) + " - message signed by submission private key")

	//submitter identity commit
	SubmitterIdenityCommit, _ := s.GetCommitMessage(ri)
	SubmitCommitAsString := fmt.Sprintf("%#v", SubmitterIdenityCommit)
	tree.Put(SubmitCommitAsString + s.userID, SubmitterIdenityCommit)
	log.Println(SubmitCommitAsString + s.userID +  " - SubmitterIdenityCommit logged")

	//paper submission commit
	PaperSubmissionCommit, _ := s.GetCommitMessagePaper(rs)
	PaperCommitAsString := fmt.Sprintf("%#v", PaperSubmissionCommit)
	tree.Put(PaperCommitAsString + strconv.Itoa(p.Id), PaperSubmissionCommit)
	log.Println(PaperCommitAsString + strconv.Itoa(p.Id) + " - PaperSubmissionCommit logged.")

	hashedIdentityCommit, _ := GetMessageHash([]byte(fmt.Sprintf("%v", s.submitterCommittedValue.CommittedValue)))
	hashedPaperCommit, _ := GetMessageHash([]byte(fmt.Sprintf("%v", s.paperCommittedValue.CommittedValue.CommittedValue)))

	SignatureSubmitterIdenityCommit, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedIdentityCommit) //Submitter Idenity Commit signed by submission private key
	tree.Put(SubmitterAsString + s.userID + "SignatureSubmitterIdentityCommit",  SignatureSubmitterIdenityCommit)
	log.Println("SignatureSubmitterIdenityCommit from userID: " + s.userID + " logged.")
	//putNextSignatureInMapSubmitter(s, signatureSubmit)

	SignaturePaperCommit, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedPaperCommit) //paper commit signed by submission private key
	tree.Put(SubmitterAsString + s.userID + "SignaturePaperCommit",  SignaturePaperCommit)
	log.Println("SignaturePaperCommit from userID: " + s.userID + " logged.")
	//putNextSignatureInMapSubmitter(s, signaturePaper)

	KsString := fmt.Sprintf("%#v", s.keys.PublicKey)
	tree.Put(KsString + s.userID, s.keys.PublicKey) //Submitters public key (Ks) is revealed to all parties
	log.Println("SubmitterPublicKey from submitter with userID: " + s.userID + " logged.") 

	hashedPaperPC, _ := GetMessageHash([]byte(fmt.Sprintf("%v", s.paperCommittedValue.CommittedValue.CommittedValue)))
	SignaturePaperCommitPC, _ := ecdsa.SignASN1(rand.Reader, pc.keys, hashedPaperPC) //PC Signs a paper commit, indicating that the paperis ready to be reviewed.
	PCsignatureAsString := fmt.Sprintf("%#v", SignaturePaperCommitPC)
	tree.Put(PCsignatureAsString + strconv.Itoa(p.Id), SignaturePaperCommitPC)
	log.Println("SignaturePaperCommitPC logged - The PC signed a paper commit.")
	//putNextSignatureInMapPC(&pc, signaturePaperPC)     

	paperList = append(paperList, *p) //List of papers, but what is it used for?? TODO

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
		Kpcr := generateSharedSecret(pc, nil, &reviewerSlice[r]) //Shared key between R and PC (Kpcr) - 
		for p := range paperSlice {

			hashedPaper, _ := GetMessageHash(EncodeToBytes(paperSlice[p])) 
			SignedPaperPC, _ := ecdsa.SignASN1(rand.Reader, pc.keys, hashedPaper)
			PaperAsString := fmt.Sprintf("%#v", paperSlice[p])
			tree.Put(PaperAsString + strconv.Itoa(paperSlice[p].Id), SignedPaperPC) 
			log.Println(PaperAsString + strconv.Itoa(paperSlice[p].Id) + "SignedPaperPC logged - The PC signed a paper")
			//putNextSignatureInMapPC(pc, pcSignature)

			encryptedPaper := Encrypt(EncodeToBytes(paperSlice[p]), Kpcr)
			encryptedAsString :=fmt.Sprintf("%#v", encryptedPaper)
			tree.Put(encryptedAsString + strconv.Itoa(paperSlice[p].Id), encryptedPaper) //Encrypted paper logged in tree
			log.Printf("\n %s %s", encryptedPaper, " encrypted paper logged")
			
			reviewerSlice[r].paperMap[p] = encryptedPaper
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
	Kpcr := generateSharedSecret(&pc, nil, r) //Shared secret key between R and PC
	tmpPaperList := []Paper{}
	for _, p := range pList {
		if p.Selected == true {
			tmpPaperList = append(tmpPaperList, p)
			putNextPaperInBidMapReviewer(r, Encrypt(EncodeToBytes(p), Kpcr))
		}
	}

	hashedBiddedPaperList, _ := GetMessageHash(EncodeToBytes(r.biddedPaperMap)) //changed from tmpPaperList to r.bidedPaperMap
	rSignature, _ := ecdsa.SignASN1(rand.Reader, r.keys, hashedBiddedPaperList)
	tree.Put(r.userID + "SignedBidByReviewer", rSignature)
	log.Printf("\n %s %s", rSignature, "Signed bid from reviewer: " + r.userID + " logged.")
	//putNextSignatureInMapReviewer(r, rSignature)

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
		reviewerAsString := fmt.Sprintf("%#v", r)
		tree.Put(reviewerAsString + "nonce", nonce) //Nonce logged
		log.Printf("\n, %s, %s", "Nonce from reviewer: " + r.userID + " - ", nonce)
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
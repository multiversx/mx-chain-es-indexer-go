package data

// AlteredAccount is a structure that holds information about an altered account
type AlteredAccount struct {
	IsSender        bool
	IsESDTOperation bool
	IsNFTOperation  bool
	TokenIdentifier string
	NFTNonce        uint64
	Type            string
}

// AlteredAccountsHandler defines the actions that an altered accounts handler should do
type AlteredAccountsHandler interface {
	Add(key string, account *AlteredAccount)
	Get(key string) ([]*AlteredAccount, bool)
	GetAll() map[string][]*AlteredAccount
	Len() int
	IsInterfaceNil() bool
}

type alteredAccounts struct {
	altered map[string][]*AlteredAccount
}

// NewAlteredAccounts will create a new instance of alteredAccounts
func NewAlteredAccounts() *alteredAccounts {
	return &alteredAccounts{
		altered: make(map[string][]*AlteredAccount),
	}
}

func (aa *alteredAccounts) Add(key string, account *AlteredAccount) {
	_, ok := aa.altered[key]
	if !ok {
		aa.altered[key] = make([]*AlteredAccount, 0)
		aa.altered[key] = append(aa.altered[key], account)

		return
	}

	isTokenOperation := account.IsESDTOperation || account.IsNFTOperation
	if !isTokenOperation {
		aa.altered[key][0].IsSender = aa.altered[key][0].IsSender || account.IsSender
		return
	}

	senderCount := 0
	for _, elem := range aa.altered[key] {
		newElementIsTokenOperation := account.IsESDTOperation || account.IsNFTOperation
		oldElementIsTokenOperation := elem.IsESDTOperation || elem.IsNFTOperation

		isSender := elem.IsSender || account.IsSender
		if isSender {
			senderCount++
		}

		shouldRewrite := newElementIsTokenOperation && !oldElementIsTokenOperation
		if shouldRewrite {
			elem.TokenIdentifier = account.TokenIdentifier
			elem.NFTNonce = account.NFTNonce
			elem.IsNFTOperation = account.IsNFTOperation
			elem.IsESDTOperation = account.IsESDTOperation
			elem.IsSender = isSender
			elem.Type = account.Type
			return
		}

		alreadyExists := elem.TokenIdentifier == account.TokenIdentifier && elem.NFTNonce == account.NFTNonce
		if alreadyExists {
			elem.IsSender = (elem.IsSender || account.IsSender) && senderCount == 1
			return
		}
	}

	if senderCount > 0 {
		// set isSender to false because regular balance change was already countered
		account.IsSender = false
	}

	aa.altered[key] = append(aa.altered[key], account)
}

func (aa *alteredAccounts) Get(key string) ([]*AlteredAccount, bool) {
	altered, ok := aa.altered[key]

	return altered, ok
}

func (aa *alteredAccounts) Len() int {
	return len(aa.altered)
}

func (aa *alteredAccounts) GetAll() map[string][]*AlteredAccount {
	if aa == nil || aa.altered == nil {
		return map[string][]*AlteredAccount{}
	}

	return aa.altered
}

// IsInterfaceNil returns true if underlying object is nil
func (aa *alteredAccounts) IsInterfaceNil() bool {
	return aa == nil
}

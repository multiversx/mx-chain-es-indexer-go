package data

// AlteredAccount is a structure that holds information about an altered account
type AlteredAccount struct {
	IsSender        bool
	IsESDTOperation bool
	IsNFTOperation  bool
	TokenIdentifier string
	NFTNonceString  string
}

type AlteredAccounts struct {
	altered map[string][]*AlteredAccount
}

func NewAlteredAccounts() *AlteredAccounts {
	return &AlteredAccounts{
		altered: make(map[string][]*AlteredAccount),
	}
}

func (aa *AlteredAccounts) Add(key string, account *AlteredAccount) {
	_, ok := aa.altered[key]
	if !ok {
		aa.altered[key] = make([]*AlteredAccount, 0)
		aa.altered[key] = append(aa.altered[key], account)

		return
	}

	isESDTOrNFT := account.IsESDTOperation || account.IsNFTOperation
	if !isESDTOrNFT {
		aa.altered[key][0].IsSender = aa.altered[key][0].IsSender || account.IsSender
		return
	}

	wasSender := false
	for _, elem := range aa.altered[key] {
		newElemIsESDT := account.IsESDTOperation || account.IsNFTOperation
		oldElemIsESDT := elem.IsESDTOperation || elem.IsNFTOperation

		wasSender = elem.IsSender || account.IsSender

		shouldRewrite := newElemIsESDT && !oldElemIsESDT
		if shouldRewrite {
			elem.TokenIdentifier = account.TokenIdentifier
			elem.NFTNonceString = account.NFTNonceString
			elem.IsNFTOperation = account.IsNFTOperation
			elem.IsESDTOperation = account.IsESDTOperation
			elem.IsSender = elem.IsSender || account.IsSender
			return
		}

		alreadyExits := elem.TokenIdentifier == account.TokenIdentifier && elem.NFTNonceString == account.NFTNonceString
		if alreadyExits {
			elem.IsSender = elem.IsSender || account.IsSender
			return
		}
	}

	if wasSender {
		// set isSender to false because regular balance change was already countered
		account.IsSender = false
	}

	aa.altered[key] = append(aa.altered[key], account)
}

func (aa *AlteredAccounts) Get(key string) ([]*AlteredAccount, bool) {
	altered, ok := aa.altered[key]

	return altered, ok
}

func (aa *AlteredAccounts) Len() int {
	return len(aa.altered)
}

func (aa *AlteredAccounts) GetAll() map[string][]*AlteredAccount {
	if aa == nil || aa.altered == nil {
		return map[string][]*AlteredAccount{}
	}

	return aa.altered
}

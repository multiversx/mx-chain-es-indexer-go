package templatesAndPolicies

// NewTemplatesAndPoliciesReader will create a new instance of templatesAndPoliciesReader
func CreateTemplatesAndPoliciesReader(useKibana bool) TemplatesAndPoliciesHandler {
	if useKibana {
		return NewTemplatesAndPolicyReaderWithKibana()
	}

	return NewTemplatesAndPolicyReaderNoKibana()
}

package templatesAndPolicies

// CreateTemplatesAndPoliciesReader will create a new instance of templatesAndPoliciesReader
func CreateTemplatesAndPoliciesReader(useKibana bool) TemplatesAndPoliciesHandler {
	if useKibana {
		return NewTemplatesAndPolicyReaderWithKibana()
	}

	return NewTemplatesAndPolicyReaderNoKibana()
}

package credentials

type unsupportedVault struct{}

func (unsupportedVault) Capabilities() Capabilities {
	return Capabilities{Name: "Sin almacén seguro", Available: false}
}
func (unsupportedVault) Get(provider string) (map[string]string, error) {
	if err := validate(provider, nil); err != nil {
		return nil, err
	}
	return map[string]string{}, nil
}
func (v unsupportedVault) Status(provider string) (Status, error) {
	if err := validate(provider, nil); err != nil {
		return Status{}, err
	}
	return Status{Provider: provider, SavedFields: []string{}, Storage: v.Capabilities()}, nil
}
func (unsupportedVault) Save(request Request) error {
	if err := validate(request.Provider, request.Fields); err != nil {
		return err
	}
	return ErrUnavailable
}
func (unsupportedVault) Delete(provider string) error {
	if err := validate(provider, nil); err != nil {
		return err
	}
	return ErrUnavailable
}

var _ Vault = unsupportedVault{}

package shopify

// GetDogDataAgeOption retrieves the "Edad" option value from the selected options
func GetDogDataAgeOption(options []SelectedOption) string {
	for _, option := range options {
		if option.Name == "Edad" {
			return option.Value
		}
	}

	return ""
}

// GetDogDataSizeOption retrieves the "Tamaño" option value from the selected options
func GetDogDataSizeOption(options []SelectedOption) string {
	for _, option := range options {
		if option.Name == "Tamaño" {
			return option.Value
		}
	}

	return ""
}

// GetDogDataConditionOption retrieves the "Condición" option value from the selected options
func GetDogDataConditionOption(options []SelectedOption) string {
	for _, option := range options {
		if option.Name == "Condición" {
			return option.Value
		}
	}

	return ""
}

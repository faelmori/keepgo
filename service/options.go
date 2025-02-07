package service

import (
	"fmt"
	"strconv"
)

// Custom type for boolean options
type BoolOption bool

// String returns the string representation of the BoolOption type.
func (b BoolOption) String() string {
	return fmt.Sprintf("%t", b)
}

// Type returns the type name of the BoolOption type.
func (b *BoolOption) Type() string {
	return "BoolOption"
}

// Set validates and sets the value for the BoolOption type.
func (b *BoolOption) Set(value string) error {
	if value == "true" || value == "false" {
		*b = (value == "true")
		return nil
	}
	return fmt.Errorf("allowed values: true, false")
}

// Custom type for string options
type StringOption string

// String returns the string representation of the StringOption type.
func (s StringOption) String() string {
	return string(s)
}

// Type returns the type name of the StringOption type.
func (s *StringOption) Type() string {
	return "StringOption"
}

// Set validates and sets the value for the StringOption type.
func (s *StringOption) Set(value string) error {
	*s = StringOption(value)
	return nil
}

// Custom type for integer options
type IntOption int

// String returns the string representation of the IntOption type.
func (i IntOption) String() string {
	return fmt.Sprintf("%d", i)
}

// Type returns the type name of the IntOption type.
func (i *IntOption) Type() string {
	return "IntOption"
}

// Set validates and sets the value for the IntOption type.
func (i *IntOption) Set(value string) error {
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid integer value: %s", value)
	}
	*i = IntOption(intValue)
	return nil
}

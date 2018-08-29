package main

// MediaType (aka MIME Type) https://www.iana.org/assignments/media-types/media-types.xhtml
type MediaType struct {
	baseType   string
	subtype    string
	parameters []string
}

// NewMediaType makes a media type with the provided values
func NewMediaType(baseType string, subtype string, parameters []string) MediaType {
	mediaType := MediaType{
		baseType:   baseType,
		subtype:    subtype,
		parameters: parameters,
	}
	return mediaType
}

// BaseType resolved
func (mediaType MediaType) BaseType() string {
	return mediaType.baseType
}

// Subtype resolved
func (mediaType MediaType) Subtype() string {
	return mediaType.subtype
}

// Parameters resolved
func (mediaType MediaType) Parameters() *[]*string {
	a := optionalStrings(mediaType.parameters)
	return &a
}

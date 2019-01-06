package conv

import (
	"github.com/abuse-mesh/abuse-mesh-go-stubs/abusemesh"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	//ErrUUIDNilPointer The given uuid pointer has a nil value
	ErrUUIDNilPointer = errors.New("Given *abusemesh.UUID is a nil value")
)

//AuuidToGuuid converts the AbuseMesh UUID used on the wire to the google/uuid format which is used in go
func AuuidToGuuid(aUUID *abusemesh.UUID) (uuid.UUID, error) {
	if aUUID == nil {
		return uuid.UUID{}, ErrUUIDNilPointer
	}
	return uuid.Parse(aUUID.Uuid)
}

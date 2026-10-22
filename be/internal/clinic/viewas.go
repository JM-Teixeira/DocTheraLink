package clinic

import "context"

// An admin "viewing as" a clinic role is treated as working in that clinic
// for the request (see auth/viewas.go). The override applies to the admin
// only: other users' clinics are resolved as usual.

type viewAsKey struct{}

type viewAs struct {
	userID uint
	clinic string
}

// WithViewAs marks ctx: userID works in clinic for this request.
func WithViewAs(ctx context.Context, userID uint, clinicID string) context.Context {
	return context.WithValue(ctx, viewAsKey{}, viewAs{userID: userID, clinic: clinicID})
}

// viewAsClinic returns the clinic userID is viewing as, if any.
func viewAsClinic(ctx context.Context, userID uint) (*string, bool) {
	v, ok := ctx.Value(viewAsKey{}).(viewAs)
	if !ok || v.userID != userID {
		return nil, false
	}
	return &v.clinic, true
}

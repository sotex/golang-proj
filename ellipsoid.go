package proj

/*
#cgo CFLAGS: -I. -I${SRCDIR}/usr/local/include
#cgo LDFLAGS: -L${SRCDIR}/usr/local/lib -lproj
#include "wrapper.h"
 */
import "C"

import (
    "fmt"
)

// Ellipsoid contains an internal object that holds everything related to a
// given ellipsoid.
type Ellipsoid struct {
    pj *C.PJ
}

// NewEllipsoid creates an ellipsoid from a WKT string or a URI.
//
func NewEllipsoid (ctx *Context, def string ) ( ell *Ellipsoid, e error ) {
    var pj *C.PJ
    pj, e = NewPJ(ctx, def, "Ellipsoid", C.PJ_CATEGORY_ELLIPSOID)
    if e == nil {
        if C.proj_get_type(pj) != C.PJ_TYPE_ELLIPSOID {
            C.proj_destroy(pj)
            pj = nil
            e = fmt.Errorf("%v does not yield an Ellipsoid", def)
            return
        }
        ell = &Ellipsoid{pj:pj}
    }
    return
}

// DestroyEllipsoid deallocate the internal ellipsoid object.
//
func (ell *Ellipsoid) DestroyEllipsoid () {
    if (*ell).pj != nil {
        C.proj_destroy((*ell).pj)
        (*ell).pj = nil
    }
}

// Handle returns the PROJ internal object to be passed to the PROJ library
// Cannot be tested against nil as it returns a pointer to a type, so use :
//   if p.HandleIsNil() { ... }
//
func (ell *Ellipsoid) Handle () (interface{}) {
    return (*ell).pj
}

// HandleIsNil returns true when the PROJ internal object is NULL.
//
func (ell *Ellipsoid)  HandleIsNil () bool {
    return (*ell).pj == (*C.PJ)(nil)
}

// TypeOf returns the ISOType of an ellipsoid (EllipsoidType).
// UnKnownType on error.
//
func (ell *Ellipsoid) TypeOf ( ) ISOType {
    return hasType(ell)
}

// SemiMajor returns the semi-major axis in meter of the given ellipsoid.
//
func (ell *Ellipsoid) SemiMajor ( ctx *Context ) ( a float64, e error ) {
    _ = C.proj_errno_reset((*ell).pj)
    var ca C.double
    // proj_ellipsoid_get_parameters fails if ell is not an ellipsoid ...
    _ = C.proj_ellipsoid_get_parameters((*ctx).pj, (*ell).pj, &ca, nil, nil, nil)
    a = float64(ca)
    return
}

// SemiMinor returns semi-minor axis in meter, whether the semi-minor is
// computed or defined of the given ellipsoid.
//
func (ell *Ellipsoid) SemiMinor ( ctx *Context ) ( b float64, bIsComputed bool, e error ) {
    _ = C.proj_errno_reset((*ell).pj)
    var cb C.double
    var cbic C.int
    // proj_ellipsoid_get_parameters fails if ell is not an ellipsoid ...
    _ = C.proj_ellipsoid_get_parameters((*ctx).pj, (*ell).pj, nil, &cb, &cbic, nil)
    b = float64(cb)
    if cbic == C.int(1) { bIsComputed = true }
    return
}

// InverseFlattening returns the inverse flattening of the given ellipsoid.
//
func (ell *Ellipsoid) InverseFlattening ( ctx *Context ) ( rf float64, e error ) {
    _ = C.proj_errno_reset((*ell).pj)
    var crf C.double
    // proj_ellipsoid_get_parameters fails if ell is not an ellipsoid ...
    _ = C.proj_ellipsoid_get_parameters((*ctx).pj, (*ell).pj, nil, nil, nil, &crf)
    rf = float64(crf)
    return
}

// Parameters returns semi-major axis in meter, semi-minor axis in meter, if
// semi-minor is computed and the inverse flattening of the given ellipsoid.
//
func (ell *Ellipsoid) Parameters ( ctx *Context ) ( a float64, b float64, bIsComputed bool, rf float64, e error ) {
    _ = C.proj_errno_reset((*ell).pj)
    var ca, cb, crf C.double
    var cbic C.int
    // proj_ellipsoid_get_parameters fails if ell is not an ellipsoid ...
    _ = C.proj_ellipsoid_get_parameters((*ctx).pj, (*ell).pj, &ca, &cb, &cbic, &crf)
    a = float64(ca)
    b = float64(cb)
    if cbic == C.int(1) { bIsComputed = true }
    rf = float64(crf)
    return
}

// Info returns information about a specific ellipsoid object.
//
func (ell *Ellipsoid) Info ( ) ( *ISOInfo ) {
    return &ISOInfo{pj:C.proj_pj_info((*ell).pj)}
}

// String returns a string representation of the ellipsoid.
//
func (ell *Ellipsoid) String ( ) string {
    return toString(ell)
}

// ProjString returns a proj-string representation of the ellipsoid.
// Empty string is returned on error.
//
func (ell *Ellipsoid) ProjString ( ctx *Context, styp StringType, opts ...string ) string {
    return toProj(ctx, ell, styp, nil)
}

// Wkt returns a WKT representation of the ellipsoid.
// Empty string is returned on error.
// `opts` can be hold the following strings :
//
//   "MULTILINE=YES" Defaults to YES, except for styp equals WKT1_ESRI
//
//   "INDENTATION_WIDTH=<number>" Defaults to 4 (when multiline output is on)
//
//   "OUTPUT_AXIS=AUTO/YES/NO" In AUTO mode, axis will be output for WKT2
//   variants, for WKT1_GDAL for ProjectedCRS with easting/northing ordering
//   (otherwise stripped), but not for WKT1_ESRI. Setting to YES will output
//   them unconditionally, and to NO will omit them unconditionally.
//
func (ell *Ellipsoid) Wkt ( ctx *Context, styp WKTType, opts ...string ) string {
    return toWkt(ctx, ell, styp, opts)
}


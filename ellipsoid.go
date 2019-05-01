package proj

/*
#cgo CFLAGS: -I. -I${SRCDIR}/usr/local/include
#cgo LDFLAGS: -L${SRCDIR}/usr/local/lib -lproj
#include "wrapper.h"
 */
import "C"

import (
    "unsafe"
    "strings"
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
    cdef := C.CString(def)
    defer C.free(unsafe.Pointer(cdef))
    switch dialect := C.proj_context_guess_wkt_dialect((*ctx).pj, cdef) ; GuessedWKTDialect(dialect) {
    case GuessedWKTUnknown  : // URI
        ac := strings.Split(def,":") // FIXME urn:ogc:def:ellipsoid::EPSG:code ?
        if len(ac) != 2 {
            e = fmt.Errorf("%v does not yield an Ellipsoid", def)
            return
        }
        cauth := C.CString(ac[0])
        defer C.free(unsafe.Pointer(cauth))
        cname := C.CString(ac[1])
        defer C.free(unsafe.Pointer(cname))
        pj = C.proj_create_from_database((*ctx).pj, cauth, cname, C.PJ_CATEGORY_ELLIPSOID, 0, nil)
        if pj == (*C.PJ)(nil) {
            e = fmt.Errorf(C.GoString(C.proj_errno_string(C.proj_context_errno((*ctx).pj))))
            return
        }
    default                 : // WKT flavor
        var ce C.PROJ_STRING_LIST
        pj = C.proj_create_from_wkt((*ctx).pj, cdef, nil, nil, &ce)
        if ce != (C.PROJ_STRING_LIST)(nil) {
            cm := C.listcat(ce)
            defer C.free(unsafe.Pointer(cm))
            defer C.proj_string_list_destroy(ce)
            e = fmt.Errorf(C.GoString(cm))
            return
        }
    }
    ell = &Ellipsoid{pj:pj}
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
    return ell.Info().Description()
}

// ProjString returns a proj-string representation of the ellipsoid.
// Empty string is returned on error.
//
func (ell *Ellipsoid) ProjString ( ctx *Context, styp StringType ) string {
    return C.GoString(C.proj_as_proj_string((*ctx).pj, (*ell).pj, C.PJ_PROJ_STRING_TYPE(styp), nil))
}

// Wkt return returns a WKT representation of the ellipsoid.
// Empty string is returned on error.
// `opts` can be :
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
    var copts **C.char
    l := len(opts)
    if l > 0 {
        copts = C.makeStringArray(C.size_t(l+1))
        for i, opt := range opts {
            copt := C.CString(opt)
            C.setStringArrayItem(copts, C.size_t(i), copt)
        }
        C.setStringArrayItem(copts, C.size_t(l), nil)
    }
    cs := C.proj_as_wkt((*ctx).pj, (*ell).pj, C.PJ_WKT_TYPE(styp), copts)
    if l > 0 {
        for i := 0 ; i < l ; i++ {
            C.free(unsafe.Pointer(C.getStringArrayItem(copts, C.size_t(i))))
        }
        C.destroyStringArray(&copts)
    }
    return C.GoString(cs)
}


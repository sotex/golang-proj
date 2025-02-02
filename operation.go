package proj

/*
#cgo CFLAGS: -I. -I${SRCDIR}/usr/local/include
#cgo LDFLAGS: -L${SRCDIR}/usr/local/lib -lproj
#include "wrapper.h"
 */
import "C"

import (
    "unsafe"
    "fmt"
)

// Operation contains an internal object that holds everything related to a given
// coordinate transformation.
//
type Operation struct {
    pj *C.PJ
}

// NewOperation creates a reference system object from a proj-string, a WKT string,
// or object code.
//
// When `bbox` is not defined then only the first element is considered :
//
//   ope, e = NewOperation(ctx, nil, "EPSG:9616")
//
//   ope, e = NewOperation(ctx, nil, "+proj=utm +zone=32 +ellps=GRS80")
//
//   ope, e = NewOperation(ctx, nil, "urn:ogc:def:coordinateOperation:EPSG::1671")
//
//   ope, e = NewOperation(ctx, nil, "WKT string")
//
// with the exception of :
//
//   ope, e = NewOperation(ctx, nil, "proj=utm", "zone=32", "ellps=GRS80")
//
// otherwise the two first elements are considered to create a transformation object
// that is a pipeline between two known coordinate reference system
// definitions.
//
//   ope, e := NewOperation(ctx, bbox, "EPSG:25832", "EPSG:25833")
//
func NewOperation ( ctx *Context, bbox *Area, def ...string ) (op *Operation, e error) {
    var pj *C.PJ
    l := len(def)
    switch {
    case l==0 :
        e = fmt.Errorf(C.GoString(C.proj_errno_string(-1)))
        return
    case l==1 :
        pj, e = NewPJ(ctx, def[0], "Operation", C.PJ_CATEGORY_COORDINATE_OPERATION)
        if e != nil { return }
    case l==2 && bbox != nil :// src and tgt CRSs
        // proj_create_crs_to_crs() is a high level function over
        // proj_create_operations() : it can then returns several operations (Cf. projinfo -s  -o PROJ -s IGNF:NTFLAMB2E.NGF84 -t IGNF:ETRS89LCC.EVRF2000)
        src, se := NewReferenceSystem(ctx, def[0])
        if se != nil { e = se ; return }
        defer src.DestroyReferenceSystem()
        tgt, te := NewReferenceSystem(ctx, def[1])
        if te != nil { e = te ; return }
        defer tgt.DestroyReferenceSystem()
        opeFactory := C.proj_create_operation_factory_context((*ctx).pj, nil)
        if opeFactory == (*C.PJ_OPERATION_FACTORY_CONTEXT)(nil) {
            e = fmt.Errorf(C.GoString(C.proj_errno_string(C.proj_context_errno((*ctx).pj))))
            return
        }
        defer C.proj_operation_factory_context_destroy(opeFactory)
        candidateOps := C.proj_create_operations((*ctx).pj, (*src).pj, (*tgt).pj, opeFactory)
        if candidateOps == (*C.PJ_OBJ_LIST)(nil) {
            e = fmt.Errorf(C.GoString(C.proj_errno_string(C.proj_context_errno((*ctx).pj))))
            return
        }
        defer C.proj_list_destroy(candidateOps)
        if C.proj_list_get_count(candidateOps) == 0 {
            e = fmt.Errorf("No operation found between '%s' and '%s'", def[0], def[1])
            return
        }
        pj = C.proj_list_get((*ctx).pj, candidateOps, C.int(0))
    default :
        defs := C.makeStringArray(C.size_t(l))
        for i, partdef := range def {
            partd := C.CString(partdef)
            C.setStringArrayItem(defs, C.size_t(i), partd)
        }
        pj = C.proj_create_argv((*ctx).pj, C.int(l), defs)
        for i := 0 ; i < l ; i++ {
            C.free(unsafe.Pointer(C.getStringArrayItem(defs,C.size_t(i))))
        }
        C.destroyStringArray(&defs)
    }
    if pj == (*C.PJ)(nil) {
        e = fmt.Errorf(C.GoString(C.proj_errno_string(C.proj_context_errno((*ctx).pj))))
        return
    }
    op = &Operation{pj:pj}
    switch op.TypeOf() {
    case Conversion,
         Transformation,
         ConcatenatedOperation,
         OtherCoordinateOperation :
        return
    default :
        e = fmt.Errorf("%v does not yield an Operation", def)
        op.DestroyOperation()
        op = nil
    }
    return
}

// DestroyOperation deallocates the internal Operation object.
//
func (op *Operation) DestroyOperation () {
    if (*op).pj != nil {
        C.proj_destroy((*op).pj)
        (*op).pj = nil
    }
}

// Handle returns the PROJ internal object to be passed to the PROJ library.
// Cannot be tested against nil as it returns a pointer to a type, so use :
//   if p.HandleIsNil() { ... }
//
func (op *Operation) Handle () (interface{}) {
    return (*op).pj
}

// HandleIsNil returns true when the PROJ internal object is NULL.
//
func (op *Operation) HandleIsNil () bool {
    return (*op).pj == (*C.PJ)(nil)
}

// TypeOf returns the ISOType of an operation (Conversion, Transformation,
// ConcatenatedOperation, OtherCoordinateOperation).
// UnKnownType on error.
//
func (op *Operation) TypeOf ( ) ISOType {
    return hasType(op)
}

func (op *Operation) fwdinv ( d Direction, aC *Coordinate ) ( aR *Coordinate, e error ) {
    var cpj, cc C.PJ_COORD
    _ = C.proj_errno_reset((*op).pj)
    // make a copy not to change coord in case of error :
    _ = C.memcpy(unsafe.Pointer(&cc), unsafe.Pointer(&((*aC).pj)), C.sizeof_PJ_COORD)
    cpj = C.proj_trans((*op).pj, C.PJ_DIRECTION(d), cc)
    if En := C.proj_errno((*op).pj) ; En != C.int(0) {
        e = fmt.Errorf(C.GoString(C.proj_errno_string(En)))
    } else {
        // everything's ok, copy back :
        _ = C.memcpy(unsafe.Pointer(&((*aC).pj)), unsafe.Pointer(&cpj), C.sizeof_PJ_COORD)
        aR = aC
    }
    return
}

func (op *Operation) fwdinv_array(d Direction, aCArray []Coordinate) (aR []Coordinate, e error) {
    ccArray := make([]C.PJ_COORD, len(aCArray))
    _ = C.proj_errno_reset((*op).pj)
    // make a copy not to change coord in case of error :
    _ = C.memcpy(unsafe.Pointer(&ccArray[0]), unsafe.Pointer(&(aCArray[0])), C.sizeof_PJ_COORD*C.size_t(len(aCArray)))
    En := C.proj_trans_array((*op).pj, C.PJ_DIRECTION(d), C.size_t(len(ccArray)), (*C.PJ_COORD)(&ccArray[0]))
    if En != C.int(0) {
        e = fmt.Errorf(C.GoString(C.proj_errno_string(En)))
    } else {
        // everything's ok, copy back :
        _ = C.memcpy(unsafe.Pointer(&(aCArray[0])), unsafe.Pointer(&ccArray[0]), C.sizeof_PJ_COORD*C.size_t(len(aCArray)))
        aR = aCArray
    }
    return
}

// Transform applies the transformation of coordinates to object
// implementing `Locatable` either from or to the CRS.
// Returns the object with transformed coordinates or nil on error.
//
func (op *Operation) Transform ( d Direction, c Locatable ) ( r Locatable, e error ) {
    xyzt := c.Location()
    xyzt, e = op.fwdinv(d, xyzt)
    if e != nil { return }
    c.SetLocation(xyzt)
    r = c
    return
}


func (op *Operation) TransformArray(d Direction, c []Coordinate) (r []Coordinate, e error) {

    c, e = op.fwdinv_array(d, c)
    r = c
    return
}

// Factors creates various cartographic properties, such as scale factors,
// angular distortion and meridian convergence.
// Depending on the underlying projection values will be calculated either
// numerically (default) or analytically.
//
// The function also calculates the partial derivatives of the given
// geographical coordinate.
//
func (op *Operation) Factors ( c *Coordinate) ( f *Factors, e error ) {
    _ = C.proj_errno_reset((*op).pj)
    var pjf C.PJ_FACTORS
    pjf = C.proj_factors((*op).pj, (*c).pj)
    if En := C.proj_errno((*op).pj) ; En != C.int(0) {
        e = fmt.Errorf(C.GoString(C.proj_errno_string(En)))
    } else {
        f = &Factors{pj:pjf}
    }
    return
}

// Info returns information about a specific operation object.
//
func (op *Operation) Info ( ) ( *ISOInfo ) {
    return &ISOInfo{pj:C.proj_pj_info((*op).pj)}
}

// String returns a string representation of the operation.
//
func (op *Operation) String ( ) string {
    return toString(op)
}

// ProjString returns a proj-string representation of the operation.
// Empty string is returned on error.
//
func (op *Operation) ProjString ( ctx *Context, styp StringType, opts ...string ) string {
    return toProj(ctx, op, styp, opts)
}

// Wkt returns a WKT representation of the operation.
// Empty string is returned on error.
// Operation can only be exported to WKT2:2018 (WKTv2r2018 or
// WKTv2r2018Simplified for `styp`).
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
func (op *Operation) Wkt ( ctx *Context, styp WKTType, opts ...string ) string {
    switch styp {
    case WKTv2r2018, WKTv2r2018Simplified :
        return toWkt(ctx, op, styp, opts)
    default :
        return ""
    }
}


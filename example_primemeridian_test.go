package proj

import (
    "fmt"
)

func ExamplePrimeMeridian () {
    s := "EPSG:8901"
    c := NewContext()
    defer c.DestroyContext()
    greenwhich, e := NewPrimeMeridian(c, s)
    if e != nil {
        fmt.Println("Ooops")
        return
    }
    greenwhichI := greenwhich.Info()
    fmt.Printf("id :%s\n", greenwhichI.ID())
    fmt.Printf("dsc:%s (%s)\n", greenwhichI.Description(), greenwhich)
    fmt.Printf("def:%s\n", greenwhichI.Definition())
    l, f, u, e := greenwhich.Parameters(c)
    if e != nil {
        fmt.Println("Ooops (Parameters)")
    }
    fmt.Printf("l :%10.2f\n", l)
    fmt.Printf("f :%10.2f\n", f)
    fmt.Printf("u :%s\n", u)
    fmt.Printf("proj-string : '%s'\n", greenwhich.ProjString(c, Version4))
    fmt.Printf("WKT : %s\n", greenwhich.Wkt(c, WKTv1GDAL, "MULTILINE=NO", "OUTPUT_AXIS=AUTO"))

    // Output:
    // id :
    // dsc:Greenwich (Greenwich)
    // def:
    // l :      0.00
    // f :      0.02
    // u :degree
    // proj-string : '+proj=noop'
    // WKT : PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]]

}


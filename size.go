// This code was started from https://github.com/DmitriyVTitov/size v1.5.0 under MIT License
package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

func calcSize(v interface{}) (uint64, error) {
	// Cache with every visited pointer so we don't count two pointers
	// to the same memory twice.
	cache := make(map[uintptr]bool)
	return sizeOf(reflect.Indirect(reflect.ValueOf(v)), cache)
}

func mustCalcSize(v interface{}) uint64 {
	s, err := calcSize(v)
	if err != nil {
		panic(err)
	}
	return s
}

func sizeOf(v reflect.Value, cache map[uintptr]bool) (uint64, error) {
	k := v.Kind()

	switch k {
	case reflect.Array:
		var sum uint64 = 0
		for i := 0; i < v.Len(); i++ {
			s, err := sizeOf(v.Index(i), cache)
			if err != nil {
				return 0, err
			}
			sum += s
		}

		return sum + uint64(v.Cap()-v.Len())*uint64(v.Type().Elem().Size()), nil
	case reflect.Slice:
		// return 0 if this node has been visited already
		if cache[v.Pointer()] {
			return 0, nil
		}
		cache[v.Pointer()] = true

		var sum uint64 = 0
		for i := 0; i < v.Len(); i++ {
			s, err := sizeOf(v.Index(i), cache)
			if err != nil {
				return 0, err
			}
			sum += s
		}

		sum += uint64(v.Cap()-v.Len()) * uint64(v.Type().Elem().Size())

		return sum + uint64(v.Type().Size()), nil
	case reflect.Struct:
		var sum uint64 = 0
		for i, n := 0, v.NumField(); i < n; i++ {
			s, err := sizeOf(v.Field(i), cache)
			if err != nil {
				return 0, err
			}
			sum += s
		}

		// Look for struct padding.
		padding := v.Type().Size()
		for i, n := 0, v.NumField(); i < n; i++ {
			padding -= v.Field(i).Type().Size()
		}

		return sum + uint64(padding), nil
	case reflect.String:
		s := v.String()
		hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
		if cache[hdr.Data] {
			return uint64(v.Type().Size()), nil
		}
		cache[hdr.Data] = true
		return uint64(len(s)) + uint64(v.Type().Size()), nil
	case reflect.Ptr:
		// return Ptr size if this node has been visited already (infinite recursion)
		if cache[v.Pointer()] {
			return uint64(v.Type().Size()), nil
		}
		cache[v.Pointer()] = true
		if v.IsNil() {
			return uint64(reflect.New(v.Type()).Type().Size()), nil
		}
		s, err := sizeOf(reflect.Indirect(v), cache)
		if err != nil {
			return 0, err
		}
		return s + uint64(v.Type().Size()), nil
	case reflect.Bool,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Int, reflect.Uint,
		reflect.Chan,
		reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Func:
		return uint64(v.Type().Size()), nil
	case reflect.Map:
		// return 0 if this node has been visited already (infinite recursion)
		if cache[v.Pointer()] {
			return 0, nil
		}
		cache[v.Pointer()] = true
		var sum uint64 = 0
		keys := v.MapKeys()
		for i := range keys {
			val := v.MapIndex(keys[i])
			// calculate size of key and value separately
			s1, err1 := sizeOf(val, cache)
			if err1 != nil {
				return 0, err1
			}
			s2, err2 := sizeOf(keys[i], cache)
			if err2 != nil {
				return 0, err2
			}
			sum += s1 + s2
		}
		// Include overhead due to unused map buckets.  10.79 comes
		// from https://golang.org/src/runtime/map.go.
		return sum + uint64(v.Type().Size()) + uint64(float64(len(keys))*10.79), nil
	case reflect.Interface:
		s, err := sizeOf(v.Elem(), cache)
		if err != nil {
			return 0, err
		}
		return s + uint64(v.Type().Size()), nil
	}

	return 0, fmt.Errorf("unknown type: %s", k.String())
}

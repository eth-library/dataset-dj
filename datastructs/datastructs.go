package datastructs

// the set struct has one attribute elements that holds true when an element is in the set
type Set struct {
	Elems map[string]bool `json:"elements"`
}

// checks if elem is contained within set s
func (s Set) Check(elem string) bool {
	return s.Elems[elem]
}

// adds elem to set s
func (s Set) Add(elem string) {
	s.Elems[elem] = true
}

// deletes elem from set s
func (s Set) Del(elem string) {
	delete(s.Elems, elem)
}

// replace the elements of a set by the contents of a slice
func (s Set) SetElemsFromSlice(slice []string) {
	s.Elems = map[string]bool{}
	for _, e := range slice {
		s.Elems[e] = true
	}
}

// return the elements of a set as a slice
func (s Set) ToSlice() []string {
	slice := []string{}
	for e := range s.Elems {
		slice = append(slice, e)
	}
	return slice
}

// create a set from a slice
func SetFromSlice(slice []string) Set {
	s := Set{Elems: map[string]bool{}}
	for _, e := range slice {
		s.Elems[e] = true
	}
	return s
}

// return a new set as the union of two sets
func SetUnion(s1, s2 Set) Set {
	newSet := Set{Elems: map[string]bool{}}
	for e := range s1.Elems {
		newSet.Elems[e] = true
	}
	for e := range s2.Elems {
		newSet.Elems[e] = true
	}
	return newSet
}

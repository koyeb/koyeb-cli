package idmapper

// Forked from https://github.com/plar/go-adaptive-radix-tree

import (
	"bytes"
	"math/bits"
	"unsafe"
)

// radixNodeKind is a node type.
type radixNodeKind int

const (
	radixLeaf    radixNodeKind = 0
	radixNode4   radixNodeKind = 1
	radixNode16  radixNodeKind = 2
	radixNode48  radixNodeKind = 3
	radixNode256 radixNodeKind = 4
)

// String returns an human friendly Kind value
func (k radixNodeKind) String() string {
	return []string{"Leaf", "Node4", "Node16", "Node48", "Node256"}[k]
}

// nolint: unused,deadcode,varcheck
const (
	node4Min = 2
	node4Max = 4

	node16Min = node4Max + 1
	node16Max = 16

	node48Min = node16Max + 1
	node48Max = 48

	node256Min = node48Max + 1
	node256Max = 256

	// Node with 48 childrens
	n48s = 6  // 2^n48s == n48m
	n48m = 64 // it should be sizeof(node48.present[0])

	// maxPrefixLen is maximum prefix length for internal nodes.
	maxPrefixLen = 10
)

type prefix [maxPrefixLen]byte

// ART node stores all available nodes, leaf and node type
type artNode struct {
	ref  unsafe.Pointer
	kind radixNodeKind
}

// a key with the null suffix will be stored as zeroChild
type node struct {
	prefixLen   uint32
	prefix      prefix
	numChildren uint16
}

// Node with 4 childrens
type node4 struct {
	node
	children [node4Max]*artNode
	keys     [node4Max]byte
	present  [node4Max]byte
}

func newNode4() *artNode {
	return &artNode{kind: radixNode4, ref: unsafe.Pointer(new(node4))}
}

// Node with 16 childrens
type node16 struct {
	node
	children [node16Max]*artNode
	keys     [node16Max]byte
	present  uint16 // need 16 bits for keys
}

func newNode16() *artNode {
	return &artNode{kind: radixNode16, ref: unsafe.Pointer(&node16{})}
}

type node48 struct {
	node
	children [node48Max]*artNode
	keys     [node256Max]byte
	present  [4]uint64 // need 256 bits for keys
}

func newNode48() *artNode {
	return &artNode{kind: radixNode48, ref: unsafe.Pointer(&node48{})}
}

// Node with 256 childrens
type node256 struct {
	node
	children [node256Max]*artNode
}

func newNode256() *artNode {
	return &artNode{kind: radixNode256, ref: unsafe.Pointer(&node256{})}
}

// Leaf node with variable key length
type leaf struct {
	key   Key
	value Value
}

func newLeaf(key Key, value Value) *artNode {
	clonedKey := make(Key, len(key))
	copy(clonedKey, key)
	return &artNode{
		kind: radixLeaf,
		ref:  unsafe.Pointer(&leaf{key: clonedKey, value: value}),
	}
}

func replaceRef(oldNode **artNode, newNode *artNode) {
	*oldNode = newNode
}

func replaceNode(oldNode *artNode, newNode *artNode) {
	*oldNode = *newNode
}

// Key Type.
// Key can be a set of any characters include unicode chars with null bytes.
type Key []byte

func (k Key) charAt(pos int) byte {
	if pos < 0 || pos >= len(k) {
		return 0
	}
	return k[pos]
}

func (k Key) valid(pos int) bool {
	return pos >= 0 && pos < len(k)
}

// Value type.
type Value interface{}

// Node interface implementation
func (an *artNode) node() *node {
	return (*node)(an.ref)
}

func (an *artNode) Kind() radixNodeKind {
	return an.kind
}

func (an *artNode) Key() Key {
	if an.isLeaf() {
		return an.leaf().key
	}

	return nil
}

func (an *artNode) isLeaf() bool {
	return an.kind == radixLeaf
}

func (an *artNode) setPrefix(key Key, prefixLen uint32) *artNode {
	node := an.node()
	node.prefixLen = prefixLen
	for i := uint32(0); i < min(prefixLen, maxPrefixLen); i++ {
		node.prefix[i] = key[i]
	}

	return an
}

func (an *artNode) matchDeep(key Key, depth uint32) uint32 /* mismatch index*/ {
	mismatchIdx := an.match(key, depth)
	if mismatchIdx < maxPrefixLen {
		return mismatchIdx
	}

	leaf := an.minimum()
	limit := min(uint32(len(leaf.key)), uint32(len(key))) - depth
	for ; mismatchIdx < limit; mismatchIdx++ {
		if leaf.key[mismatchIdx+depth] != key[mismatchIdx+depth] {
			break
		}
	}

	return mismatchIdx
}

// Find the minimum leaf under a artNode
func (an *artNode) minimum() *leaf {
	switch an.kind {
	case radixLeaf:
		return an.leaf()

	case radixNode4:
		node := an.node4()
		return node.children[0].minimum()

	case radixNode16:
		node := an.node16()
		return node.children[0].minimum()

	case radixNode48:
		node := an.node48()
		idx := uint8(0)
		for node.present[idx>>n48s]&(1<<(idx%n48m)) == 0 {
			idx++
		}
		if node.children[node.keys[idx]] != nil {
			return node.children[node.keys[idx]].minimum()
		}

	case radixNode256:
		node := an.node256()
		if len(node.children) > 0 {
			idx := 0
			for ; node.children[idx] == nil; idx++ {
				// find 1st non empty
			}
			return node.children[idx].minimum()
		}
	}

	return nil // that should never happen in normal case
}

func (an *artNode) index(c byte) int {
	switch an.kind {
	case radixNode4:
		node := an.node4()
		for idx := 0; idx < int(node.numChildren); idx++ {
			if node.keys[idx] == c {
				return idx
			}
		}

	case radixNode16:
		node := an.node16()
		bitfield := uint(0)
		for i := uint(0); i < node16Max; i++ {
			if node.keys[i] == c {
				bitfield |= (1 << i)
			}
		}
		mask := (1 << node.numChildren) - 1
		bitfield &= uint(mask)
		if bitfield != 0 {
			return bits.TrailingZeros(bitfield)
		}

	case radixNode48:
		node := an.node48()
		if s := node.present[c>>n48s] & (1 << (c % n48m)); s > 0 {
			if idx := int(node.keys[c]); idx >= 0 {
				return idx
			}
		}

	case radixNode256:
		return int(c)
	}

	return -1 // not found
}

var nodeNotFound *artNode

func (an *artNode) findChild(c byte, valid bool) **artNode {
	idx := an.index(c)
	if idx != -1 {
		switch an.kind {
		case radixNode4:
			return &an.node4().children[idx]

		case radixNode16:
			return &an.node16().children[idx]

		case radixNode48:
			return &an.node48().children[idx]

		case radixNode256:
			return &an.node256().children[idx]
		}
	}

	return &nodeNotFound
}

func (an *artNode) node4() *node4 {
	return (*node4)(an.ref)
}

func (an *artNode) node16() *node16 {
	return (*node16)(an.ref)
}

func (an *artNode) node48() *node48 {
	return (*node48)(an.ref)
}

func (an *artNode) node256() *node256 {
	return (*node256)(an.ref)
}

func (an *artNode) leaf() *leaf {
	return (*leaf)(an.ref)
}

func (an *artNode) addChild4(c byte, valid bool, child *artNode) bool {
	node := an.node4()

	// grow to node16
	if node.numChildren >= node4Max {
		newNode := an.grow()
		newNode.addChild(c, valid, child)
		replaceNode(an, newNode)
		return true
	}

	// just add a new child
	i := uint16(0)
	for ; i < node.numChildren; i++ {
		if c < node.keys[i] {
			break
		}
	}

	limit := node.numChildren - i
	for j := limit; limit > 0 && j > 0; j-- {
		node.keys[i+j] = node.keys[i+j-1]
		node.present[i+j] = node.present[i+j-1]
		node.children[i+j] = node.children[i+j-1]
	}
	node.keys[i] = c
	node.present[i] = 1
	node.children[i] = child
	node.numChildren++
	return false
}

func (an *artNode) addChild16(c byte, valid bool, child *artNode) bool {
	node := an.node16()

	if node.numChildren >= node16Max {
		newNode := an.grow()
		newNode.addChild(c, valid, child)
		replaceNode(an, newNode)
		return true
	}

	idx := node.numChildren
	bitfield := uint(0)
	for i := uint(0); i < node16Max; i++ {
		if node.keys[i] > c {
			bitfield |= (1 << i)
		}
	}
	mask := (1 << node.numChildren) - 1
	bitfield &= uint(mask)
	if bitfield != 0 {
		idx = uint16(bits.TrailingZeros(bitfield))
	}

	for i := node.numChildren; i > idx; i-- {
		node.keys[i] = node.keys[i-1]
		node.present = (node.present & ^(1 << i)) | ((node.present & (1 << (i - 1))) << 1)
		node.children[i] = node.children[i-1]
	}

	node.keys[idx] = c
	node.present |= (1 << idx)
	node.children[idx] = child
	node.numChildren++
	return false
}

func (an *artNode) addChild48(c byte, valid bool, child *artNode) bool {
	node := an.node48()
	if node.numChildren >= node48Max {
		newNode := an.grow()
		newNode.addChild(c, valid, child)
		replaceNode(an, newNode)
		return true
	}

	index := byte(0)
	for node.children[index] != nil {
		index++
	}

	node.keys[c] = index
	node.present[c>>n48s] |= (1 << (c % n48m))
	node.children[index] = child
	node.numChildren++
	return false
}

func (an *artNode) addChild256(c byte, valid bool, child *artNode) bool {
	node := an.node256()
	node.numChildren++
	node.children[c] = child

	return false
}

func (an *artNode) addChild(c byte, valid bool, child *artNode) bool {
	switch an.kind {
	case radixNode4:
		return an.addChild4(c, valid, child)

	case radixNode16:
		return an.addChild16(c, valid, child)

	case radixNode48:
		return an.addChild48(c, valid, child)

	case radixNode256:
		return an.addChild256(c, valid, child)
	}

	return false
}

func (an *artNode) copyMeta(src *artNode) *artNode {
	if src == nil {
		return an
	}

	d := an.node()
	s := src.node()

	d.numChildren = s.numChildren
	d.prefixLen = s.prefixLen

	for i, limit := uint32(0), min(s.prefixLen, maxPrefixLen); i < limit; i++ {
		d.prefix[i] = s.prefix[i]
	}

	return an
}

func (an *artNode) grow() *artNode {
	switch an.kind {
	case radixNode4:
		node := newNode16().copyMeta(an)

		d := node.node16()
		s := an.node4()

		for i := uint16(0); i < s.numChildren; i++ {
			if s.present[i] != 0 {
				d.keys[i] = s.keys[i]
				d.present |= (1 << i)
				d.children[i] = s.children[i]
			}
		}

		return node

	case radixNode16:
		node := newNode48().copyMeta(an)

		d := node.node48()
		s := an.node16()

		var numChildren byte
		for i := uint16(0); i < s.numChildren; i++ {
			if s.present&(1<<i) != 0 {
				ch := s.keys[i]
				d.keys[ch] = numChildren
				d.present[ch>>n48s] |= (1 << (ch % n48m))
				d.children[numChildren] = s.children[i]
				numChildren++
			}
		}

		return node

	case radixNode48:
		node := newNode256().copyMeta(an)

		d := node.node256()
		s := an.node48()

		for i := uint16(0); i < node256Max; i++ {
			if s.present[i>>n48s]&(1<<(i%n48m)) != 0 {
				d.children[i] = s.children[s.keys[i]]
			}
		}

		return node
	}

	return nil
}

// Leaf methods
func (l *leaf) match(key Key) bool {
	if key == nil || len(l.key) != len(key) {
		return false
	}

	return bytes.Equal(l.key[:len(key)], key)
}

// Base node methods
func (an *artNode) match(key Key, depth uint32) uint32 /* 1st mismatch index*/ {
	idx := uint32(0)
	if len(key)-int(depth) < 0 {
		return idx
	}

	node := an.node()

	limit := min(min(node.prefixLen, maxPrefixLen), uint32(len(key))-depth)
	for ; idx < limit; idx++ {
		if node.prefix[idx] != key[idx+depth] {
			return idx
		}
	}

	return idx
}

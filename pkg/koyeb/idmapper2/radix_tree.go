package idmapper2

// Forked from https://github.com/plar/go-adaptive-radix-tree

// Callback function type for tree traversal.
type Callback func(key Key, value Value)

// RadixTree is an Adaptive Radix Tree implementation for our internal Mapper.
type RadixTree struct {
	root *artNode
	size int
}

// NewRadixTree creates a new adaptive radix tree.
func NewRadixTree() *RadixTree {
	return &RadixTree{}
}

// Insert inserts given key into the radix tree.
// You can add duplicate keys, this radix tree is like a set.
func (t *RadixTree) Insert(key Key, value Value) {
	ok := t.recursiveInsert(&t.root, key, value, 0)
	if ok {
		t.size++
	}
}

// Size returns the number of keys stored in the tree.
func (t *RadixTree) Size() int {
	if t == nil || t.root == nil {
		return 0
	}

	return t.size
}

// MinimalLength computes the minimal key length that avoid collision if keys are shorten/compressed.
func (t *RadixTree) MinimalLength(minimum int) int {
	if t == nil || t.root == nil {
		return minimum
	}

	v, ok := t.recursiveMinimalLength(t.root, 0)
	if ok && int(v) > minimum {
		return int(v)
	}

	return minimum
}

// ForEach executes a provided callback once per leaf node by default.
func (t *RadixTree) ForEach(callback Callback) {
	t.recursiveForEachCallback(t.root, callback)
}

// String returns an human friendly format of the adaptive radix tree.
func (t *RadixTree) String() string {
	return debugNode(t.root)
}

func (t *RadixTree) recursiveMinimalLength(current *artNode, depth uint32) (uint32, bool) {
	if current == nil {
		return 0, false
	}

	switch current.kind {
	case radixLeaf:
		return depth, true

	case radixNode4:
		node := current.node4()
		nlist := node.children[:]
		ndepth := depth + node.prefixLen + 1

		return t.forEachChildrenMinimalLength(nlist, ndepth)

	case radixNode16:
		node := current.node16()
		nlist := node.children[:]
		ndepth := depth + node.prefixLen + 1

		return t.forEachChildrenMinimalLength(nlist, ndepth)

	case radixNode48:
		node := current.node48()
		nlist := node.children[:]
		ndepth := depth + node.prefixLen + 1

		return t.forEachChildrenMinimalLength(nlist, ndepth)

	case radixNode256:
		node := current.node256()
		nlist := node.children[:]
		ndepth := depth + node.prefixLen + 1

		return t.forEachChildrenMinimalLength(nlist, ndepth)

	default:
		return 0, false
	}
}

func (t *RadixTree) forEachChildrenMinimalLength(childrens []*artNode, depth uint32) (uint32, bool) {
	lengthVal := uint32(0)
	lengthOk := false

	for _, child := range childrens {
		if child != nil {
			val, ok := t.recursiveMinimalLength(child, depth)
			if ok && val > lengthVal {
				lengthVal = val
				lengthOk = ok
			}
		}
	}

	return lengthVal, lengthOk
}

func (t *RadixTree) recursiveForEachCallback(current *artNode, callback Callback) {
	if current == nil {
		return
	}

	switch current.kind {
	case radixLeaf:
		node := current.leaf()
		callback(node.key, node.value)

	case radixNode4:
		list := current.node4().children[:]
		t.forEachChildrenCallback(list, callback)

	case radixNode16:
		list := current.node16().children[:]
		t.forEachChildrenCallback(list, callback)

	case radixNode48:
		list := current.node48().children[:]
		t.forEachChildrenCallback(list, callback)

	case radixNode256:
		list := current.node256().children[:]
		t.forEachChildrenCallback(list, callback)
	}
}

func (t *RadixTree) forEachChildrenCallback(childrens []*artNode, callback Callback) {
	for _, child := range childrens {
		if child != nil {
			t.recursiveForEachCallback(child, callback)
		}
	}
}

func (t *RadixTree) recursiveInsert(curNode **artNode, key Key, value Value, depth uint32) bool {
	current := *curNode
	if current == nil {
		replaceRef(curNode, newLeaf(key, value))
		return true
	}

	if current.isLeaf() {
		return t.recursiveInsertOnLeaf(curNode, current, key, value, depth)
	}

	node := current.node()
	if node.prefixLen == 0 {
		return t.recursiveInsertOnNextNode(curNode, current, key, value, depth)
	}

	return t.recursiveInsertOnCurrentNode(curNode, current, key, value, depth)
}

func (t *RadixTree) recursiveInsertOnLeaf(curNode **artNode, current *artNode, key Key, value Value, depth uint32) bool {
	leaf := current.leaf()
	if leaf.match(key) {
		return false
	}

	// new value, split the leaf into new node4
	newLeaf := newLeaf(key, value)
	leaf2 := newLeaf.leaf()
	leafsLCP := t.longestCommonPrefix(leaf, leaf2, depth)

	newNode := newNode4()
	newNode.setPrefix(key[depth:], leafsLCP)
	depth += leafsLCP

	newNode.addChild(leaf.key.charAt(int(depth)), leaf.key.valid(int(depth)), current)
	newNode.addChild(leaf2.key.charAt(int(depth)), leaf2.key.valid(int(depth)), newLeaf)
	replaceRef(curNode, newNode)

	return true
}

func (t *RadixTree) recursiveInsertOnCurrentNode(curNode **artNode, current *artNode, key Key, value Value, depth uint32) bool {
	node := current.node()

	prefixMismatchIdx := current.matchDeep(key, depth)
	if prefixMismatchIdx >= node.prefixLen {
		depth += node.prefixLen
		return t.recursiveInsertOnNextNode(curNode, current, key, value, depth)
	}

	newNode := newNode4()
	node4 := newNode.node()
	node4.prefixLen = prefixMismatchIdx
	for i := 0; i < int(min(prefixMismatchIdx, maxPrefixLen)); i++ {
		node4.prefix[i] = node.prefix[i]
	}

	if node.prefixLen <= maxPrefixLen {
		node.prefixLen -= (prefixMismatchIdx + 1)
		newNode.addChild(node.prefix[prefixMismatchIdx], true, current)

		for i, limit := uint32(0), min(node.prefixLen, maxPrefixLen); i < limit; i++ {
			node.prefix[i] = node.prefix[prefixMismatchIdx+i+1]
		}

	} else {
		node.prefixLen -= (prefixMismatchIdx + 1)
		leaf := current.minimum()
		newNode.addChild(leaf.key.charAt(int(depth+prefixMismatchIdx)), leaf.key.valid(int(depth+prefixMismatchIdx)), current)

		for i, limit := uint32(0), min(node.prefixLen, maxPrefixLen); i < limit; i++ {
			node.prefix[i] = leaf.key[depth+prefixMismatchIdx+i+1]
		}
	}

	// Insert the new leaf
	newNode.addChild(key.charAt(int(depth+prefixMismatchIdx)), key.valid(int(depth+prefixMismatchIdx)), newLeaf(key, value))
	replaceRef(curNode, newNode)

	return true
}

func (t *RadixTree) recursiveInsertOnNextNode(curNode **artNode, current *artNode, key Key, value Value, depth uint32) bool {
	// Find a child to recursive to
	next := current.findChild(key.charAt(int(depth)), key.valid(int(depth)))
	if *next != nil {
		return t.recursiveInsert(next, key, value, depth+1)
	}

	// No Child, artNode goes with us
	current.addChild(key.charAt(int(depth)), key.valid(int(depth)), newLeaf(key, value))

	return true
}

func (t *RadixTree) longestCommonPrefix(l1 *leaf, l2 *leaf, depth uint32) uint32 {
	l1key, l2key := l1.key, l2.key
	idx, limit := depth, min(uint32(len(l1key)), uint32(len(l2key)))
	for ; idx < limit; idx++ {
		if l1key[idx] != l2key[idx] {
			break
		}
	}

	return idx - depth
}

// X helpers
func min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

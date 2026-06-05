package sensitive

// Filter 敏感词过滤器（DFA 算法）
type Filter struct {
	root *node
}

type node struct {
	children map[rune]*node
	isEnd    bool
}

// NewFilter 创建过滤器
func NewFilter() *Filter {
	return &Filter{
		root: &node{children: make(map[rune]*node)},
	}
}

// AddWord 添加敏感词
func (f *Filter) AddWord(word string) {
	current := f.root
	for _, char := range word {
		if next, ok := current.children[char]; ok {
			current = next
		} else {
			newNode := &node{children: make(map[rune]*node)}
			current.children[char] = newNode
			current = newNode
		}
	}
	current.isEnd = true
}

// AddWords 批量添加敏感词
func (f *Filter) AddWords(words []string) {
	for _, word := range words {
		f.AddWord(word)
	}
}

// Filter 过滤敏感词（替换为 *）
func (f *Filter) Filter(text string) string {
	runes := []rune(text)
	length := len(runes)
	result := make([]rune, length)
	copy(result, runes)

	for i := 0; i < length; i++ {
		current := f.root
		j := i
		lastMatch := -1

		for j < length {
			if next, ok := current.children[runes[j]]; ok {
				current = next
				if current.isEnd {
					lastMatch = j
				}
				j++
			} else {
				break
			}
		}

		if lastMatch != -1 {
			for k := i; k <= lastMatch; k++ {
				result[k] = '*'
			}
			i = lastMatch
		}
	}

	return string(result)
}

// Contains 检查是否包含敏感词
func (f *Filter) Contains(text string) bool {
	runes := []rune(text)
	length := len(runes)

	for i := 0; i < length; i++ {
		current := f.root
		j := i

		for j < length {
			if next, ok := current.children[runes[j]]; ok {
				current = next
				if current.isEnd {
					return true
				}
				j++
			} else {
				break
			}
		}
	}

	return false
}

// FindAll 查找所有敏感词
func (f *Filter) FindAll(text string) []string {
	runes := []rune(text)
	length := len(runes)
	var found []string

	for i := 0; i < length; i++ {
		current := f.root
		j := i

		for j < length {
			if next, ok := current.children[runes[j]]; ok {
				current = next
				if current.isEnd {
					found = append(found, string(runes[i:j+1]))
				}
				j++
			} else {
				break
			}
		}
	}

	return found
}

// ReplaceWith 替换敏感词为指定字符
func (f *Filter) ReplaceWith(text string, replaceChar rune) string {
	runes := []rune(text)
	length := len(runes)
	result := make([]rune, length)
	copy(result, runes)

	for i := 0; i < length; i++ {
		current := f.root
		j := i
		lastMatch := -1

		for j < length {
			if next, ok := current.children[runes[j]]; ok {
				current = next
				if current.isEnd {
					lastMatch = j
				}
				j++
			} else {
				break
			}
		}

		if lastMatch != -1 {
			for k := i; k <= lastMatch; k++ {
				result[k] = replaceChar
			}
			i = lastMatch
		}
	}

	return string(result)
}

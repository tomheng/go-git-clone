package delta

import (
	"io"

	"github.com/tomheng/gogit/git"
)

const (
	CopySection = iota
	InsertSection
)

//parse copy or insert section info from delta reader
func ParseCopyOrInsert(r io.Reader) (stype int, offset, length int64, err error) {
	b, err := git.ReadOneByte(r)
	if err != nil {
		return
	}
	var _b byte
	switch git.IsMsbSet(b) {
	case true: //copy section
		stype = CopySection
		//check last 4 byte
		for i := uint(0); i < 7; i++ {
			//we should read 1 byte from reader
			_b = 0
			if b&(1<<i) != 0 {
				_b, err = git.ReadOneByte(r)
				if err != nil {
					break
				}
			}
			//fmt.Printf("i:%d, _b:%b, :%b\n", i, _b, b&(1<<i))
			if i < 4 {
				offset += int64(_b) << (i * 8)
			} else {
				length += int64(_b) << ((i - 4) * 8)
			}
		}
	case false: //insert section
		stype = InsertSection
		length = int64(b)
	}
	return
}

func Patch(base io.SectionReader, delta io.Reader) (target io.ReadWriter, err error) {
	baseLen, err := git.ParseVarLen(delta)
	if err != nil {
		return
	}
	targetLen, err := git.ParseVarLen(delta)
	_ = baseLen
	_ = targetLen
	if err != nil {
		return
	}
	for {
		st, offset, length, err := ParseCopyOrInsert(delta)
		if err != nil {
			break
		}
		bs := make([]byte, length)
		switch st {
		case CopySection:
			_, err := base.ReadAt(bs, offset)
			if err != nil {
				break
			}
		case InsertSection:
			_, err := delta.Read(bs)
			if err != nil {
				break
			}
		}
		_, err = target.Write(bs)
		if err != nil {
			break
		}
		/*if n1 != n2 {
			err = errors.New("read not equal to write")
			break;
		}*/
	}
	//fmt.Printf("bl:%d,tl:%d", baseLen, targetLen)
	return
}

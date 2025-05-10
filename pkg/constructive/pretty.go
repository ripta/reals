package constructive

import (
	"fmt"
	"strings"
)

func Pretty(c Real) string {
	sb := &strings.Builder{}
	pretty(sb, c)
	return sb.String()
}

func pretty(sb *strings.Builder, c Real) {
	if c == nil {
		sb.WriteString("nil")
		return
	}

	switch v := c.(type) {
	case *named:
		sb.WriteString(fmt.Sprintf("Named(%q)", v.Name))
	case *constructiveInteger:
		sb.WriteString(fmt.Sprintf("Integer(%s)", v.i))
	case *constructiveMultiplication:
		sb.WriteString("Multiply(")
		pretty(sb, v.a)
		sb.WriteString(", ")
		pretty(sb, v.b)
		sb.WriteString(")")
	case *constructiveAddition:
		pretty(sb, v.a)
		sb.WriteString(" + ")
		pretty(sb, v.b)
	case *constructiveMultiplicativeInverse:
		sb.WriteString("Inverse(")
		pretty(sb, v.r)
		sb.WriteString(")")
	default:
		sb.WriteString(fmt.Sprintf("%T %+v", v, v))
	}
}

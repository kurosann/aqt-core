package dec

import (
	"reflect"

	"github.com/shopspring/decimal"

	"github.com/kurosann/aqt-core/logger"
)

func Int[T I](i T) decimal.Decimal {
	return decimal.NewFromInt(int64(i))
}
func Float[T F](f T) decimal.Decimal {
	return decimal.NewFromFloat(float64(f))
}
func String(s string) decimal.Decimal {
	if s == "" {
		return Zero
	}
	return decimal.RequireFromString(s)
}

var Zero = decimal.Zero

func AddMulti[T Num](i1 T, i2 ...T) decimal.Decimal {
	total := N(i1)
	for _, v := range i2 {
		total = total.Add(N(v))
	}
	return total
}
func SubMulti[T Num](i1 T, i2 ...T) decimal.Decimal {
	total := N(i1)
	for _, v := range i2 {
		total = total.Sub(N(v))
	}
	return total
}
func PowMulti[T Num, R Num](i1 T, i2 ...T) decimal.Decimal {
	total := N(i1)
	for _, v := range i2 {
		total = total.Pow(N(v))
	}
	return total
}
func MulMulti[T Num](i1 T, i2 ...T) decimal.Decimal {
	total := N(i1)
	for _, v := range i2 {
		total = total.Mul(N(v))
	}
	return total
}
func DivMulti[T Num](i1 T, i2 ...T) decimal.Decimal {
	total := N(i1)
	for _, v := range i2 {
		total = total.Div(N(v))
	}
	return total
}
func Add[T Num, R Num](i1 T, i2 R) decimal.Decimal {
	return N(i1).Add(N(i2))
}
func Sub[T Num, R Num](i1 T, i2 R) decimal.Decimal {
	return N(i1).Sub(N(i2))
}
func Pow[T Num, R Num](i1 T, i2 R) decimal.Decimal {
	return N(i1).Pow(N(i2))
}
func Mul[T Num, R Num](i1 T, i2 R) decimal.Decimal {
	return N(i1).Mul(N(i2))
}
func Div[T Num, R Num](i1 T, i2 R) decimal.Decimal {
	return N(i1).Div(N(i2))
}

const LoopCount = 50

func Sqrt(i1 decimal.Decimal) decimal.Decimal {
	x := Div(i1, 3)
	lastX := Int(0)
	for i := 0; i < LoopCount; i++ {
		x = i1.Div(Mul(x, 2)).Add(Div(x, 2))
		if Eq(x, lastX) {
			break
		}
		lastX = x
	}
	return x
}

// Eq 等于
func Eq[T Num, R Num](i1 T, i2 R) bool {
	return N(i1).Equal(N(i2))
}

// GT 大于
func GT[T Num, R Num](i1 T, i2 R) bool {
	return N(i1).GreaterThan(N(i2))
}

// GTE 大于等于
func GTE[T Num, R Num](i1 T, i2 R) bool {
	return N(i1).GreaterThanOrEqual(N(i2))
}

// LT 小于
func LT[T Num, R Num](i1 T, i2 R) bool {
	return N(i1).LessThan(N(i2))
}

// LTE 小于等于
func LTE[T Num, R Num](i1 T, i2 R) bool {
	return N(i1).LessThanOrEqual(N(i2))
}

func Sum(in []decimal.Decimal) decimal.Decimal {
	total := N(0)
	for _, a := range in {
		total = Add(total, a)
	}
	return total
}
func Avg(in []decimal.Decimal) decimal.Decimal {
	if len(in) == 0 {
		return Sum(in)
	}
	return Div(Sum(in), len(in))
}

func N(e any) decimal.Decimal {
	switch t := e.(type) {
	case int64:
		return Int(t)
	case int32:
		return Int(t)
	case int8:
		return Int(t)
	case int16:
		return Int(t)
	case byte:
		return Int(t)
	case int:
		return Int(t)
	case float64:
		return Float(t)
	case float32:
		return Float(t)
	case string:
		return String(t)
	case decimal.Decimal:
		return t
	default:
		ofA := reflect.ValueOf(e)
		if !ofA.IsValid() || ofA.IsNil() || ofA.IsZero() {
			return Zero
		}
		if s, ok := e.(interface{ String() string }); ok {
			return String(s.String())
		}
		logger.Panicf("decimal type is not support: %v", t)
		return Int(0)
	}
}

// RoundSignificant 保留5位小数
func RoundSignificant(d decimal.Decimal) decimal.Decimal {
	// 如果数字大于等于1，直接保留5位小数
	if GTE(d, 1) {
		return d.Round(5)
	}

	// 对于小于1的数字，从第一个非零数字开始计算
	str := d.String()

	foundDecimal := false
	zeroCount := 0

	for _, c := range str {
		if c == '.' {
			foundDecimal = true
			continue
		}
		if foundDecimal {
			if c == '0' {
				zeroCount++
			} else {
				zeroCount++ // 包括第一个非零数字
				break
			}
		}
	}

	// 计算需要保留的小数位数
	precision := 5 + zeroCount
	//logger.Warnf("计算需要保留的小数位数%d,保留后：%v", precision, d.Round(int32(precision)))
	return d.Round(int32(precision))
}

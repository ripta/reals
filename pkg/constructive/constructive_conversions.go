package constructive

func DegreesToRadians(degrees Real) Real {
	// π / 180 * degrees
	return Multiply(Divide(Pi(), FromInt(180)), degrees)
}

func RadiansToDegrees(radians Real) Real {
	// 180 / π * radians
	return Multiply(Divide(FromInt(180), Pi()), radians)
}

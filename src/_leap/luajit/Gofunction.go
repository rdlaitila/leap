package luajit

// A Gofunction is a Go function that may be registered with the Lua
// interpreter and called by Lua programs.
//
// In order to communicate properly with Lua, a Go function must use the
// following protocol, which defines the way parameters and results are
// passed: a Go function receives its arguments from Lua in its stack
// in direct order (the first argument is pushed first). So, when the
// function starts, s.Gettop returns the number of arguments received by the
// function. The first argument (if any) is at index 1 and its last argument
// is at index s.Gettop. To return values to Lua, a Go function just pushes
// them onto the stack, in direct order (the first result is pushed first),
// and returns the number of results. Any other value in the stack below
// the results will be properly discarded by Lua. Like a Lua function,
// a Go function called by Lua can also return many results.
//
// As an example, the following function receives a variable number of
// numerical arguments and returns their average and sum:
//
// 	func foo(s *luajit.State) int {
// 		n := s.Gettop()		// number of arguments
// 		sum := 0.0
// 		for i := 1; i <= n; i++ {
// 			if !s.Isnumber(i) {
// 				s.Pushstring("incorrect argument")
// 				s.Error()
// 			}
// 			sum += s.Tonumber(i)
// 		}
// 		s.Pushnumber(sum/n)	// first result
// 		s.Pushnumber(sum)	// second result
// 		return 2		// number of results
// 	}
type Gofunction func(*State) int
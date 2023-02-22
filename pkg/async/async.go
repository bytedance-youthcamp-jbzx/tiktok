package async

import "context"

// Future interface has the method signature for await

type Future interface {
	Await() interface{}
}

type future struct {
	await func(ctx context.Context) interface{}
}

func (f future) Await() interface{} {
	return f.await(context.Background())
}

// Exec executes the async function
func Exec(f func() interface{}) Future {
	var result interface{}
	c := make(chan struct{})

	go func() {
		defer close(c)
		result = f()
	}()

	return future{
		await: func(ctx context.Context) interface{} {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-c:
				return result
			}
		},
	}

}

// Copyright 2021 Amadeus s.a.s
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package async provide different kind of interface and implementation to manipulate a bit more safely the go routing.
//
// First things  provided by this package is an implementation of the pattern Async/Await you can find in Javascript
// It should be used when you need to run multiple asynchronous task and wait for each of them to finish.
// Example:
//  func doneAsync() int {
//		// asynchronous task
//		time.Sleep(3 * time.Second)
//		return 1
//	}
//
//  func synchronousTask() {
//  	next := async.Async(func() interface{} {
//			return doneAsync()
//  	})
//		// do some other stuff
//  	// then wait for the end of the asynchronous task and get back the result
//  	result := next.Await()
//  	// do something with the result
//  }
//
// It is useful to use this implementation when you want to paralyze quickly some short function like paralyzing multiple HTTP request.
// You definitely won't use this implementation if you want to create a cron or a long task. Instead you should implement the interface SimpleTask or Task for doing that.
package async

import "context"

type Future interface {
	Await() interface{}
	AwaitWithContext(ctx context.Context) interface{}
}

type next struct {
	await func(ctx context.Context) interface{}
}

func (n *next) Await() interface{} {
	return n.await(context.Background())
}

func (n *next) AwaitWithContext(ctx context.Context) interface{} {
	return n.await(ctx)
}

// Async executes the asynchronous function
func Async(f func() interface{}) Future {
	var result interface{}
	// c is going to be used to catch only the signal when the channel is closed.
	c := make(chan struct{})
	go func() {
		defer close(c)
		result = f()
	}()
	return &next{
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

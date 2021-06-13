# Checkpoint

This is a small lib which extends the standard Go error wrapping functionality with `fmt.Errorf(...%w..., err)`.

## Motivation
First you have to read this https://blog.golang.org/go1.13-errors and decide yourself if you want to use the checkpoints 
or if you just stick to `fmt.Errorf` because in many cases it may be enough.


Often I get errors without knowing where exactly what happened.
Using the default Go `fmt.Errorf` functionality with `%w` is a good way to add additional data to an error.

But I have some problems with that:
1. It is only possible to wrap one error at a time with %w.  
    I would like to use a pattern like this:
    ```go
    var (
        ErrAnErrorWithDescription = errors.New("a super description")
        ...
    )
    
    func AnyFunction() error {
        ...
        err := AFunctionReturningAnError()
        if err != nil {
            return fmt.Errorf("%v:\n%w", ErrAnErrorWithDescription, err)
        }
        ...
        return nil
    }
    ```
    Now this works great, but I am only able to check for `errors.Is(err, wantErr)` with wantErr being the error from
    `AFunctionReturningAnError()`. It is not possible to check for `errors.Is(err, ErrAnErrorWithDescription)`.
    I would like to check and get both errors using `errors.Is` / `errors.As` which is not easily possible.  
    Especially in tests it can be useful to be able to check for both, the error which was thrown in `AnyFunction` but also the error from
    `AFunctionReturningAnError`
2. It is very annoying to always need to format the error string again. If I for example leave out the \n, I get one long error message line 
    which is not readable. That's why this lib just uses the same formatting for everything (as far as possible).  
    Also it includes some additional information this way, like the source code line and filename.
3. Often it can be very helpful if the error logs contain a code path instead of only one simple error. Yes, it may 
    get very big and full stack traces like in Java are not always that helpful.  
    But this shouldn't be a problem here, because the checkpoints only record the places where a checkpoint is used.  
    That way you have full control over if you just want to re-return the error or if you want to create a new checkpoint.
   
## Usage
There are two ways to create a new checkpoint:
```go
checkpoint.From(err)
```
and
```go
checkpoint.Wrap(err, ErrAnErrorWithDescription)
```
For the exact usage and behaviour just consult the documentation `go doc -all`.
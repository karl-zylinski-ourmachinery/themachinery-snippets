# Update implementation function pointers

Goal: If we remove an implementation and re-add (from hot-reload), the function pointers in any
existing implementation should be updated.

Currently we just store the "pointers" to the implementation, so in this sequence:

~~~c
reg->add_implementation("test_i", test_version, &impl_a);
test_i *x = reg->first_implementation("test_i", test_version)
reg->remove_implementation("test_i", test_version, &impl_a);
reg->add_implementation("test_i", test_version, &impl_b);
~~~

`x` will still point to `impl_a` and if we call `x->f()` it will call the old implementation in
`impl_a`. We want to fix it so that it calls the implementation in `impl_b` instead.

## Option 1: `add_implementation()` copies the implementation to a fixed memory location

We can follow what we already do for APIs and instead of having the API registry store pointers
to the implementations, we copy the implementations to an internal buffer. If we add back the
same implementation again, it overwrites that same memory buffer which will update its function
pointers.

The code might look something like this:

~~~c
reg->add_implementation("test_i", test_version, &impl_a, sizeof(impl_a), "impl_1");
test_i *x = reg->first_implementation("test_i", test_version)
reg->remove_implementation("test_i", test_version, &impl_a), sizeof(impl_a), "impl_1";
reg->add_implementation("test_i", test_version, &impl_b, sizeof(impl_b), "impl_1");
~~~

With this approach `x` will point to the internal API registry copy of `impl_a`, when we call
`add_implementation()` with `impl_b`, this internal copy will get overwritten by `impl_b`, so after
this, any call to `x->f()` will call the implementation in `imbpl_b`.

Note that since we know need to copy the implementation data we need to know its size. Also,
since there can be multiple implementations of the same interface we need a unique identifier to
tell us which implementation we are "updating" with the second `add_implementation()`. In the
example above, we use a simple string `"impl_1"` for this purpose.

For the `tm_add_or_remove_implementation()` macro, we might be able to get rid of the identifier
by using `__FILE__##__COUNTER__` to identify the implementation.

Another consequence of this is that since we need to know the size of the things we are copying,
we can no longer use functions as interfaces (because functions don't have a size), we have to use
function *pointers* instead. So instead of this:

~~~c
typedef void tm_the_truth_create_types_i(tm_the_truth_o *tt);

tm_add_or_remove_implementation(reg, load, tm_the_truth_create_types_i, truth__create);
~~~

We will have to do this:

~~~c
typedef void (*tm_the_truth_create_types_i)(tm_the_truth_o *tt);

static tm_the_truth_create_types_i truth_create_i = truth__create
tm_add_or_remove_implementation(reg, load, tm_the_truth_create_types_i, &truth__create_i, "impl_1");
~~~

## Option 2: You copy the struct to a fixed memory location before calling `add_implementation()`

Instead of Option 1, we can continue to let the API registry just keep track of pointers and
instead make it the responsibility of the caller to copy the pointer to a fixed memory location
before calling `add_implementation()`. It might look something like this:

~~~c
test_i *impl = reg->static_variable(TM_STATIC_HASH("impl_1"), sizeof(impl_a), "impl_1");
memcpy(impl, &impl_a, sizeof(impl_a));
reg->add_implementation("test_i", test_version, impl);
test_i *x = reg->first_implementation("test_i", test_version)
reg->remove_implementation("test_i", test_version, impl);
memcpy(impl, &impl_b, sizeof(impl_b));
reg->add_implementation("test_i", test_version, impl);
~~~

With macros, we might be able to simplify this somewhat:

~~~c
tm_add_or_remove_implementation(reg, load, test_i, tm_to_static(impl_a));
~~~

Again, if we want to use this for function interfaces, we need a level of indirection (because
if the returned implementation pointer points straight to the function, there is no way of
replacing it), so we need something like:

~~~c
typedef void (*tm_the_truth_create_types_i)(tm_the_truth_o *tt);

tm_add_or_remove_implementation(reg, load, test_i, tm_to_static_f(truth__create));
~~~

Here, a separate macro `tm_to_static_f` is used to convert the function pointer.


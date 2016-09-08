local ret = run_command "echo -n foo"

if ret:stdout() ~= "foo" then
    error("it should get 'foo' but" .. ret:stdout())
end

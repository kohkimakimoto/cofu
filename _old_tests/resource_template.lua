template "/tmp/testing_template" {
    mode = "0644",
    owner = "root",
    group = "root",
    content = [=[
hoge
]=],
    verify = {
        "grep 'hoge' /tmp/testing_template",
    },
}

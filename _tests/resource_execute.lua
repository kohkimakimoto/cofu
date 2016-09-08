-- execute command.
execute "echo 'hoge' > /tmp/testing_execute" {
    verify = {
        "grep 'hoge' /tmp/testing_execute",
    },
}

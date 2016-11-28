-- create directory
directory "/tmp/testing_directory" {
    action = "create",
    mode = "0755",
    owner = "root",
    group = "root",
    verify = {
        "test -d /tmp/testing_directory",
        "test -x /tmp/testing_directory",
        "ls -ld  /tmp/testing_directory | awk '{print $3, $4 }' | grep 'root root'",
    },
}

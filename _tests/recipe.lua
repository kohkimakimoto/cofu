-- user "kohkimakimoto" {
--     uid = "1010",
--     shell = "/bin/bash",
-- }
--
-- group "testgroup1980" {
--     gid = "1980",
-- }
--
-- group "testgroup1981" {
--     gid = 1981,
-- }
--
-- execute "ls -la" {
--
-- }
--
-- software_package "httpd" {
--     action = "install",
-- }
--
-- service "httpd" {
--     action = {"enable", "start"},
-- }


execute "ls" {

}

local result = run_command("echo -n Hello")
print(result:stdout())
print(result:stderr())
print(result:combined())
print(result:exit_status())
print(result:failure())
print(result:success())

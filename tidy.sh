go mod tidy

for mod in $(find examples testdata -name go.mod); do
  dir=$(dirname $mod)
  (cd $dir && go mod tidy)
done

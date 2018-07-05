# Refresh "model"
rm -rf ./vendor/github.com/lagoon-platform/model/*.go
cp ../model/*.go  ./vendor/github.com/lagoon-platform/model/
rm -rf ./vendor/github.com/lagoon-platform/model/params/*.go
cp ../model/params/*.go  ./vendor/github.com/lagoon-platform/model/params/

# Refresh "engine"
rm -rf ./vendor/github.com/lagoon-platform/engine/*.go
cp ../engine/*.go  ./vendor/github.com/lagoon-platform/engine/
rm -rf ./vendor/github.com/lagoon-platform/engine/ssh/*.go
cp ../engine/ssh/*.go  ./vendor/github.com/lagoon-platform/engine/ssh

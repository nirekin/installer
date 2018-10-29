echo "INSTALLER copydep..."

# Refresh "model"
rm -rf ./vendor/github.com/ekara-platform/model/*.go
mkdir ./vendor/github.com/ekara-platform/model/
cp ../model/*.go  ./vendor/github.com/ekara-platform/model/

# Refresh "engine"
rm -rf ./vendor/github.com/ekara-platform/engine/*.go
mkdir ./vendor/github.com/ekara-platform/engine/
cp ../engine/*.go  ./vendor/github.com/ekara-platform/engine/

rm -rf ./vendor/github.com/ekara-platform/engine/ansible/*.go
mkdir ./vendor/github.com/ekara-platform/engine/ansible/
cp ../engine/ansible/*.go  ./vendor/github.com/ekara-platform/engine/ansible/

rm -rf ./vendor/github.com/ekara-platform/engine/component/*.go
mkdir ./vendor/github.com/ekara-platform/engine/component/
cp ../engine/component/*.go  ./vendor/github.com/ekara-platform/engine/component/

rm -rf ./vendor/github.com/ekara-platform/engine/ssh/*.go
mkdir ./vendor/github.com/ekara-platform/engine/ssh/
cp ../engine/ssh/*.go  ./vendor/github.com/ekara-platform/engine/ssh/

rm -rf ./vendor/github.com/ekara-platform/engine/util/*.go
mkdir ./vendor/github.com/ekara-platform/engine/util/
cp ../engine/util/*.go  ./vendor/github.com/ekara-platform/engine/util/


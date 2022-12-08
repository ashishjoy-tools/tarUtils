VERSION=$1

git tag $VERSION
git push origin $VERSION
GOPROXY=proxy.golang.org go list -m  "github.com/ashishjoy-tools/tarUtils@$VERSION"

package youyouayedee

import "fmt"

// DCEDomain indicates the type of identifier stored in a DCE-based UUID.
type DCEDomain byte

const (
	Person DCEDomain = 0x00
	Group  DCEDomain = 0x01
	Org    DCEDomain = 0x02
)

type DCEDomainData struct {
	GoName string
	Name   string
}

var dceDomainDataArray = [...]DCEDomainData{
	{GoName: "Person", Name: "person"},
	{GoName: "Group", Name: "group"},
	{GoName: "Org", Name: "organization"},
}

func (domain DCEDomain) IsValid() bool {
	p := uint(domain)
	q := uint(len(dceDomainDataArray))
	return p < q
}

func (domain DCEDomain) Data() DCEDomainData {
	p := uint(domain)
	q := uint(len(dceDomainDataArray))
	if p < q {
		return dceDomainDataArray[p]
	}
	goName := fmt.Sprintf("youyouayedee.DCEDomain(%d)", p)
	name := fmt.Sprintf("<unknown DCE domain %d>", p)
	return DCEDomainData{GoName: goName, Name: name}
}

func (domain DCEDomain) GoString() string {
	return domain.Data().GoName
}

func (domain DCEDomain) String() string {
	return domain.Data().Name
}

var (
	_ fmt.GoStringer = DCEDomain(0)
	_ fmt.Stringer   = DCEDomain(0)
)

package infrastructureutilitytransactor

import (
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	"github.com/MateWorkspace/mate-golang/pgxdt"
)

func NewPgxdtImpl(dt pgxdt.Transactor) domaincontractsutility.Transactor {
	return dt
}

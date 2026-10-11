package compositionmain

import (
	"github.com/Masterminds/squirrel"
	"github.com/MateWorkspace/lockmate/internal/config"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	infrastructurerepositoryclassicauth "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/classic_auth"
	infrastructurerepositoryclient "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/client"
	infrastructurerepositoryclientorigin "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/client_origin"
	infrastructurerepositorygoogleauth "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/google_auth"
	infrastructurerepositorypermission "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/permission"
	infrastructurerepositoryrole "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/role"
	infrastructurerepositoryrolepermission "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/role_permission"
	infrastructurerepositoryspace "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/space"
	infrastructurerepositoryspacemember "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/space_member"
	infrastructurerepositoryspacememberrole "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/space_member_role"
	infrastructurerepositoryuser "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/user"
	infrastructureutilitykeygen "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/keygen"
	infrastructureutilitylogger "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/logger"
	infrastructureutilitypassword "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/password"
	infrastructureutilitypermissionguard "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/permission_guard"
	infrastructureutilitytoken "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/token"
	infrastructureutilitytracer "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/tracer"
	infrastructureutilitytransactor "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/transactor"
	infrastructureutilityvalidator "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/validator"
	"github.com/MateWorkspace/mate-golang/pgxdt"
)

type infrastructure struct {
	space           domaincontractsrepository.Space
	permission      domaincontractsrepository.Permission
	role            domaincontractsrepository.Role
	user            domaincontractsrepository.User
	classicAuth     domaincontractsrepository.ClassicAuth
	googleAuth      domaincontractsrepository.GoogleAuth
	client          domaincontractsrepository.Client
	clientOrigin    domaincontractsrepository.ClientOrigin
	spaceMember     domaincontractsrepository.SpaceMember
	rolePermission  domaincontractsrepository.RolePermission
	spaceMemberRole domaincontractsrepository.SpaceMemberRole
	tracer          domaincontractsutility.Tracer
	logger          domaincontractsutility.Logger
	keygen          domaincontractsutility.Keygen
	password        domaincontractsutility.Password
	transactor      domaincontractsutility.Transactor
	token           domaincontractsutility.Token
	validator       domaincontractsutility.Validator
	permissionGuard domaincontractsutility.PermissionGuard
}

func (l *launcher) newInfrastructure() {
	question := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
	dollar := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	dt := pgxdt.NewPgxdt(l.drv.postgresPool)
	tracer := infrastructureutilitytracer.NewContextImpl()

	l.infra = &infrastructure{
		space:           infrastructurerepositoryspace.NewPostgresImpl(dt, &question, &dollar),
		permission:      infrastructurerepositorypermission.NewPostgresImpl(dt, &question, &dollar),
		role:            infrastructurerepositoryrole.NewPostgresImpl(dt, &question, &dollar),
		user:            infrastructurerepositoryuser.NewPostgresImpl(dt, &question, &dollar),
		classicAuth:     infrastructurerepositoryclassicauth.NewPostgresImpl(dt, &question, &dollar),
		googleAuth:      infrastructurerepositorygoogleauth.NewPostgresImpl(dt, &question, &dollar),
		client:          infrastructurerepositoryclient.NewPostgresImpl(dt, &question, &dollar),
		clientOrigin:    infrastructurerepositoryclientorigin.NewPostgresImpl(dt, &question, &dollar),
		spaceMember:     infrastructurerepositoryspacemember.NewPostgresImpl(dt, &question, &dollar),
		rolePermission:  infrastructurerepositoryrolepermission.NewPostgresImpl(dt, &question, &dollar),
		spaceMemberRole: infrastructurerepositoryspacememberrole.NewPostgresImpl(dt, &question, &dollar),
		tracer:          tracer,
		logger:          infrastructureutilitylogger.NewSlogImpl(l.drv.slogLogger, tracer, config.Env.LoggerLevel, config.Env.LoggerUseTrace),
		keygen:          infrastructureutilitykeygen.NewGeneratorImpl(),
		password:        infrastructureutilitypassword.NewBcryptImpl(config.Env.PasswordBcryptCost),
		transactor:      infrastructureutilitytransactor.NewPgxdtImpl(pgxdt.NewTransactor(l.drv.postgresPool)),
		token:           infrastructureutilitytoken.NewJwtImpl(config.Env.TokenAccessSecret, config.Env.TokenRefreshSecret, config.Env.TokenAccessDuration, config.Env.TokenRefreshDuration),
		validator:       infrastructureutilityvalidator.NewValidatorImpl(),
		permissionGuard: infrastructureutilitypermissionguard.NewClaimsImpl(),
	}
}

import { Component } from '@angular/core';
import { MatLegacyDialog as MatDialog } from '@angular/material/legacy-dialog';
import { ActivatedRoute, Params, Router } from '@angular/router';
import { BehaviorSubject, from, Observable, of } from 'rxjs';
import { catchError, finalize, map, take } from 'rxjs/operators';
import { CreationType, MemberCreateDialogComponent } from 'src/app/modules/add-member-dialog/member-create-dialog.component';
import { PolicyComponentServiceType } from 'src/app/modules/policies/policy-component-types.enum';
import {
  BRANDING,
  COMPLEXITY,
  DOMAIN,
  GENERAL,
  IDP,
  LOCKOUT,
  LOGIN,
  LOGINTEXTS,
  MESSAGETEXTS,
  NOTIFICATIONS,
  PRIVACYPOLICY,
  SECRETS,
  SECURITY,
  OIDC,
} from 'src/app/modules/settings-list/settings';
import { SidenavSetting } from 'src/app/modules/sidenav/sidenav.component';
import { InstanceDetail, State } from 'src/app/proto/generated/zitadel/instance_pb';
import { Member } from 'src/app/proto/generated/zitadel/member_pb';
import { User } from 'src/app/proto/generated/zitadel/user_pb';
import { AdminService } from 'src/app/services/admin.service';
import { Breadcrumb, BreadcrumbService, BreadcrumbType } from 'src/app/services/breadcrumb.service';
import { ToastService } from 'src/app/services/toast.service';

@Component({
  selector: 'cnsl-instance',
  templateUrl: './instance.component.html',
  styleUrls: ['./instance.component.scss'],
})
export class InstanceComponent {
  public instance?: InstanceDetail.AsObject;
  public PolicyComponentServiceType: any = PolicyComponentServiceType;
  private loadingSubject: BehaviorSubject<boolean> = new BehaviorSubject<boolean>(false);
  public loading$: Observable<boolean> = this.loadingSubject.asObservable();
  public totalMemberResult: number = 0;
  public membersSubject: BehaviorSubject<Member.AsObject[]> = new BehaviorSubject<Member.AsObject[]>([]);
  public State: any = State;

  public id: string = '';
  public settingsList: SidenavSetting[] = [
    GENERAL,
    // notifications
    // { showWarn: true, ...NOTIFICATIONS },
    NOTIFICATIONS,
    // login
    LOGIN,
    IDP,
    COMPLEXITY,
    LOCKOUT,

    DOMAIN,
    // appearance
    BRANDING,
    MESSAGETEXTS,
    LOGINTEXTS,
    // others
    PRIVACYPOLICY,
    OIDC,
    SECRETS,
    SECURITY,
  ];
  constructor(
    public adminService: AdminService,
    private dialog: MatDialog,
    private toast: ToastService,
    breadcrumbService: BreadcrumbService,
    private router: Router,
    activatedRoute: ActivatedRoute,
  ) {
    this.loadMembers();

    const instanceBread = new Breadcrumb({
      type: BreadcrumbType.INSTANCE,
      name: 'Instance',
      routerLink: ['/instance'],
      hideNav: true,
    });

    breadcrumbService.setBreadcrumb([instanceBread]);

    this.adminService
      .getMyInstance()
      .then((instanceResp) => {
        if (instanceResp.instance) {
          this.instance = instanceResp.instance;
        }
      })
      .catch((error) => {
        this.toast.showError(error);
      });

    const breadcrumbs = [
      new Breadcrumb({
        type: BreadcrumbType.INSTANCE,
        name: 'Instance',
        routerLink: ['/instance'],
      }),
    ];
    breadcrumbService.setBreadcrumb(breadcrumbs);

    activatedRoute.queryParams.pipe(take(1)).subscribe((params: Params) => {
      const { id } = params;
      if (id) {
        this.id = id;
      }
    });
  }

  public loadMembers(): void {
    this.loadingSubject.next(true);
    from(this.adminService.listIAMMembers(100, 0))
      .pipe(
        map((resp) => {
          if (resp.details?.totalResult) {
            this.totalMemberResult = resp.details.totalResult;
          } else {
            this.totalMemberResult = 0;
          }
          return resp.resultList;
        }),
        catchError(() => of([])),
        finalize(() => this.loadingSubject.next(false)),
      )
      .subscribe((members) => {
        this.membersSubject.next(members);
      });
  }

  public openAddMember(): void {
    const dialogRef = this.dialog.open(MemberCreateDialogComponent, {
      data: {
        creationType: CreationType.IAM,
      },
      width: '400px',
    });

    dialogRef.afterClosed().subscribe((resp) => {
      if (resp) {
        const users: User.AsObject[] = resp.users;
        const roles: string[] = resp.roles;

        if (users && users.length && roles && roles.length) {
          Promise.all(
            users.map((user) => {
              return this.adminService.addIAMMember(user.id, roles);
            }),
          )
            .then(() => {
              this.toast.showInfo('IAM.TOAST.MEMBERADDED', true);
              setTimeout(() => {
                this.loadMembers();
              }, 1000);
            })
            .catch((error) => {
              this.toast.showError(error);
              setTimeout(() => {
                this.loadMembers();
              }, 1000);
            });
        }
      }
    });
  }

  public showDetail(): void {
    this.router.navigate(['/instance', 'members']);
  }
}

import { Component, OnInit } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { AppInitService } from "../../../app.init.service";
import { Member, Project, Role } from "../../../project/project";
import { Subject } from 'rxjs/Subject';
import { SharedService } from "../../shared.service";
import { MessageService } from "../../message-service/message.service";
import { Observable } from "rxjs/Observable";
import { AlertType } from "../../shared.types";
import { CsModalChildBase } from "../../cs-modal-base/cs-modal-child-base";

@Component({
  selector: 'project-member',
  templateUrl: './member.component.html'
})
export class MemberComponent extends CsModalChildBase implements OnInit {
  currentUser: {[key: string]: any};
  role: Role = new Role();
  availableMembers: Member[];
  selectedMember: Member = new Member();
  project: Project = new Project();
  isLeftPane: boolean;
  isRightPane: boolean;
  doSet: boolean;
  doUnset: boolean;
  memberSubject: Subject<Member[]> = new Subject<Member[]>();
  isActionWip: boolean = false;

  constructor(private sharedService: SharedService,
              private messageService: MessageService,
              private appInitService: AppInitService,
              private translateService: TranslateService) {
    super();
  }
  
  ngOnInit(): void {
    this.currentUser = this.appInitService.currentUser;
  }

  retrieveAvailableMembers() {
    this.sharedService.getProjectMembers(this.project.project_id).then((members: Array<Member>) => {
      this.sharedService.getAvailableMembers().then((availableMembers: Array<Member>) => {
        this.availableMembers = availableMembers;
        this.availableMembers.forEach((am: Member) => {
          members.forEach((member: Member) => {
            if (member.project_member_user_id === am.project_member_user_id) {
              am.project_member_id = member.project_member_id;
              am.project_member_role_id = member.project_member_role_id;
              am.isMember = true;
              }
          });
        });
        this.memberSubject.subscribe((changedMembers: Array<Member>) => {
          this.availableMembers = changedMembers;
          });
        });
      });
  }

  openMemberModal(project: Project): Observable<string> {
    this.project = project;
    this.role.role_id = 1;
    this.retrieveAvailableMembers();
    return super.openModal();
  }

  pickUpMember(member: Member) {
    this.selectedMember = member;
    this.doSet = false;
    this.doUnset = false;
    let isProjectOwner = (this.project.project_owner_id === this.currentUser.user_id);
    let isSelf = (this.currentUser.user_id === this.selectedMember.project_member_user_id);
    let isSystemAdmin = (this.currentUser.user_system_admin === 1);
    let isOnesProject = (this.project.project_owner_id === this.selectedMember.project_member_user_id);
    if((isSelf && isProjectOwner) || (isSystemAdmin && isOnesProject)) {
      this.doSet = false;
      this.doUnset = false;
    } else { 
      if(isProjectOwner || isSystemAdmin) {
        this.doSet = this.isLeftPane;
        this.doUnset = this.isRightPane;
      }
    }
    this.role.role_id = this.selectedMember.project_member_role_id;
  }

  pickUpRole(role: Role) {
    this.selectedMember.project_member_role_id = role.role_id;
    this.isActionWip = true;
    this.sharedService.addOrUpdateProjectMember(this.project.project_id,
        this.selectedMember.project_member_user_id, 
        this.selectedMember.project_member_role_id)
      .then(() => this.displayInlineMessage('PROJECT.SUCCESSFUL_CHANGED_MEMBER_ROLE', 'alert-info', [this.selectedMember.project_member_username]))
      .catch(() => this.displayInlineMessage('PROJECT.FAILED_TO_CHANGE_MEMBER_ROLE', 'alert-danger'));
  }

  setMember(): void {
    this.availableMembers.forEach((member: Member) => {
      if (member.project_member_user_id === this.selectedMember.project_member_user_id) {
        member.project_member_role_id = this.role.role_id;
        this.isActionWip = true;
        this.sharedService.addOrUpdateProjectMember(this.project.project_id,
            this.selectedMember.project_member_user_id, 
            this.selectedMember.project_member_role_id)
          .then(() => this.displayInlineMessage('PROJECT.SUCCESSFUL_ADDED_MEMBER', 'alert-info', [this.selectedMember.project_member_username]))
          .catch(() => this.displayInlineMessage('PROJECT.FAILED_TO_ADD_MEMBER', 'alert-danger'));
        member.isMember = true;
      }
    });
    this.memberSubject.next(this.availableMembers);
  }

  unsetMember(): void {
    this.availableMembers.forEach((member: Member) => {
      if (member.project_member_user_id === this.selectedMember.project_member_user_id) {
        this.selectedMember.project_member_id = 1;
        member.isMember = false;
        this.isActionWip = true;
        this.sharedService
          .deleteProjectMember(this.project.project_id, this.selectedMember.project_member_user_id)
          .then(() => this.displayInlineMessage('PROJECT.SUCCESSFUL_REMOVED_MEMBER', 'alert-info', [this.selectedMember.project_member_username]))
          .catch(() => this.displayInlineMessage('PROJECT.FAILED_TO_REMOVE_MEMBER', 'alert-danger'));
      }
    });
    this.memberSubject.next(this.availableMembers);
  }

  displayInlineMessage(message: string, alertType: AlertType, params?: object): void {
    this.translateService.get(message, params).subscribe((res: string) => {
      this.isActionWip = false;
      this.messageService.showAlert(res, {alertType: alertType, view: this.alertView})
    });
  }
}
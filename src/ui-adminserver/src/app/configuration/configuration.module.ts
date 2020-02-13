import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CfgCardsComponent } from './cfg-cards.component';
import { RouterModule } from '@angular/router';
import { ApiserverComponent } from './apiserver/apiserver.component';
import { GogitsComponent } from './gogits/gogits.component';
import { JenkinsComponent } from './jenkins/jenkins.component';
import { KvmComponent } from './kvm/kvm.component';
import { LdapComponent } from './ldap/ldap.component';
import { EmailComponent } from './email/email.component';
import { OthersComponent } from './others/others.component';
import { ToTopDirective } from '../shared/to-top.directive';
import { ClarityModule } from '@clr/angular';
import { BoardComponentsLibraryModule } from 'board-components-library';
import { CfgCardsService } from './cfg-cards.service';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { SharedModule } from '../shared/shared.module';
import { TranslateModule } from '@ngx-translate/core';

@NgModule({
  declarations: [
    CfgCardsComponent,
    ApiserverComponent,
    GogitsComponent,
    JenkinsComponent,
    KvmComponent,
    LdapComponent,
    EmailComponent,
    OthersComponent,
    ToTopDirective
  ],
  providers: [
    CfgCardsService,
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    CommonModule,
    ClarityModule,
    BoardComponentsLibraryModule,
    SharedModule,
    TranslateModule,
    RouterModule.forChild([{ path: '', component: CfgCardsComponent }])
  ]
})
export class ConfigurationModule { }

import { Injectable } from '@angular/core';
import { 
  CanActivate, CanActivateChild, Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot
 }                    from '@angular/router';
import { AppInitService } from '../app.init.service';
import { Message } from './message-service/message';
import { MessageService } from './message-service/message.service';

@Injectable()
export class AuthGuard implements CanActivate, CanActivateChild {
  
  constructor(
    private appInitService: AppInitService,
    private messageService: MessageService,
    private router: Router
  ){}

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return new Promise<boolean>((resolve, reject)=>{
      this.appInitService
        .currentUser()
        .then(res=>resolve(true))
        .catch(err=>{
          this.router.navigate(['/sign-in']);
          resolve(false);
        });
    });
  }

  canActivateChild(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return this.canActivate(route, state);
  }
}
import { Component, EventEmitter, Input, OnInit, Output, ViewContainerRef } from '@angular/core';
import { ClrDatagridStateInterface } from '@clr/angular';
import { NodeService } from '../node.service';
import { CsModalChildBase } from '../../shared/cs-modal-base/cs-modal-child-base';
import { NodeControlStatus, NodeStatus, ServiceInstance } from '../node.types';

@Component({
  selector: 'app-node-service-control',
  templateUrl: './node-service-control.component.html',
  styleUrls: ['./node-service-control.component.css']
})
export class NodeServiceControlComponent extends CsModalChildBase implements OnInit {
  @Input() nodeCurrent: NodeStatus;
  @Input() instanceCount: number;
  @Input() deletable: boolean;
  @Output() instanceCountChange: EventEmitter<number>;
  @Output() deletableChange: EventEmitter<boolean>;
  nodeControlStatus: NodeControlStatus;
  serviceInstanceList: Array<ServiceInstance>;
  curPageIndex = 1;
  curPageSize = 6;

  constructor(private nodeService: NodeService,
              private view: ViewContainerRef) {
    super();
    this.nodeControlStatus = new NodeControlStatus({});
    this.serviceInstanceList = Array<ServiceInstance>();
    this.instanceCountChange = new EventEmitter<number>();
    this.deletableChange = new EventEmitter<boolean>();
  }

  ngOnInit() {
    this.refreshData();
  }

  get phaseStyle(): { [p: string]: string } {
    switch (this.nodeControlStatus.nodePhase) {
      case 'Pending':
        return {color: 'darkorange'};
      case 'Running':
        return {color: 'green'};
      case 'Terminal':
        return {color: 'red'};
      default:
        return {color: 'black'};
    }
  }

  refreshData() {
    this.nodeService.getNodeControlStatus(this.nodeCurrent.nodeName).subscribe(
      (res: NodeControlStatus) => {
        this.nodeControlStatus = res;
        this.instanceCountChange.emit(this.nodeControlStatus.serviceInstances.length);
        this.deletableChange.emit(this.nodeControlStatus.deletable);
        this.curPageIndex = 1;
        this.retrieve({page: {from: 0, to: 5}});
      }
    );
  }

  retrieve(page: ClrDatagridStateInterface) {
    if (Reflect.has(page, 'page')) {
      const from = page.page.from;
      const to = page.page.to + 1;
      this.serviceInstanceList = this.nodeControlStatus.serviceInstances.slice(from, to);
    }
  }


}

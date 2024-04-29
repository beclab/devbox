import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
//import { CreateApplicationConfig } from '@devbox/core';
//import { Container } from '../../core/src/index';

//const _db = level('./data');
//const ApplicationKey = 'applications';

const appcfg = {
  version: 'v0.1',
  metadata: {
    name: 'app name',
    icon: "app's icon",
    description: 'app description',
    appid: 'app id',
    title: 'app title',
    version: 'app version',
    categories: ['dev'],
    target: 'new',
  },
  entrances: [
    {
      name: 'entrance name',
      host: 'entrance service name',
      port: 80,
      icon: 'entrance icon',
      title: 'entrance title',
      authLevel: 'private',
    },
  ],
  spec: {
    versionName: 'version name',
    fullDescription: 'long story',
    upgradeDescription: 'upgrade',
    promoteImage: ['img1 url', 'img2 url'],
    promoteVideo: 'video url',
    subCategory: 'tools',
    developer: 'developer',
    requiredMemory: '0.5Gi',
    requiredDisk: '500Mi',
    supportClient: {
      edge: 'edge plugin download url',
      android: 'android app download url',
      ios: 'ios app download url',
      windows: 'windows client download url',
      mac: 'mac client download url',
      linux: 'linux client download url',
    },
    requiredGpu: '8G',
    requiredCpu: '500m',
  },
  permission: {
    appData: true,
    sysData: [
      {
        group: 'group name',
        dataType: 'data type',
        version: 'version',
        ops: ['Create', 'List'],
      },
    ],
  },
  middleware: {
    postgres: {
      username: 'pg user name',
      password: 'pg password',
      databases: [
        {
          name: 'db1',
          distributed: false,
        },
      ],
    },
    redis: {
      password: 'redis password',
      databases: [
        {
          name: 'namespace1',
        },
        {
          name: 'namespace2',
        },
      ],
    },

    mongodb: {
      username: 'mongo user name',
      password: 'mongo password',
      databases: [
        {
          name: 'db1',
        },
        {
          name: 'db2',
        },
      ],
    },

    zincSearch: {
      username: 'zinc user name',
      password: 'zinc password',
      indexes: [
        {
          name: 'index name',
        },
      ],
    },
  },

  options: {
    policies: [
      {
        entranceName: 'entrance name',
        description: 'policy description',
        uriRegex: 'uri regex',
        level: 'two_factor',
        oneTime: false,
        validDuration: '5s',
      },
    ],

    analytics: {
      enabled: false,
    },

    dependencies: [
      {
        name: 'app name',
        version: '>=0.1.0',
        type: 'application',
      },
    ],

    appScope: {
      clusterScoped: true,
      appRef: ['client1', 'client2'],
    },

    websocket: {
      port: 80,
      url: '/path/to/callback',
    },
  },
};

let myc = 1;

@Injectable()
export class AppService implements OnModuleInit {
  private readonly logger = new Logger(AppService.name);

  public apps: any = [];
  public cfgs: Record<string, any> = {};
  public appContainers: Record<string, any> = {};
  public myContainers = [];
  constructor() {
    //
  }

  async onModuleInit(): Promise<void> {
    // try {
    //   const res = await _db.get(ApplicationKey);
    //   this.apps = JSON.parse(res);
    //   this.logger.log('init apps', this.apps);
    // } catch (e) {}
  }

  async saveApp(config: any): Promise<any> {
    const app = {
      id: this.apps.length,
      appName: config.name,
      devEnv: config.devEnv,
      createTime: new Date().toISOString(),
      updateTime: new Date().toISOString(),
      chart: '/app',
      entrance: '1ass12sc.xx.snowinning.com',
      ide: '1ass12sc.xx.snowinning.com/proxy/3000',
    };

    this.apps.push(app);
    // await _db.put(ApplicationKey, JSON.stringify(this.apps));
    return { appId: app.id };
  }

  async removeApp(name: string) {
    this.apps = this.apps.filter((app) => app.appName !== name);
  }

  async getCfg(app_name: string): Promise<any> {
    if (app_name in this.cfgs) {
      return this.cfgs[app_name];
    }
    return appcfg;
  }

  async setCfg(app_name: string, cfg: any): Promise<any> {
    this.cfgs[app_name] = cfg;
  }

  async getAppContainers(app_name: string): Promise<any> {
    if (app_name in this.appContainers) {
      return this.appContainers[app_name];
    } else {
      this.appContainers[app_name] = [
        {
          image: 'image1',
          podSelector: 'app=devapp1, label1=value1',
          containerName: 'container1',
        },
        {
          image: 'image2',
          podSelector: 'app=devapp2, label1=value2',
          containerName: 'container2',
        },
      ];
      return this.appContainers[app_name];
    }
  }

  async getMyContainers(): Promise<any> {
    return this.myContainers;
  }

  async bindContainer({
    containerId,
    appId,
    podSelector,
    containerName,
    devEnv,
  }: {
    containerId?: number;
    appId: number;
    podSelector: string;
    containerName: string;
    devEnv?: string;
  }): Promise<any> {
    const app = this.apps.find((app) => app.id === appId);
    if (!app) {
      throw new Error(`app ${appId} not found`);
    }

    const container = this.appContainers[app.appName].find(
      (container) =>
        container.podSelector == podSelector &&
        container.containerName == containerName,
    );
    if (!container) {
      throw new Error(`container ${containerName} ${podSelector} not found`);
    }

    if (containerId) {
      const index = this.myContainers.findIndex(
        (container) => container.id == containerId,
      );
      if (index < 0) {
        throw new Error(`container ${containerId} not found`);
      }
      console.log(index);
      this.myContainers[index].podSelector = podSelector;
      this.myContainers[index].containerName = containerName;
      this.myContainers[index].appId = appId;
      this.myContainers[index].createTime = new Date().toISOString();
      this.myContainers[index].updateTime = new Date().toISOString();
      this.myContainers[index].state = 'Running';
      this.myContainers[index].appName = app.appName;
      this.myContainers[index].devEnv = devEnv;
    } else {
      const my: any = {};
      my.id = myc++;
      my.podSelector = podSelector;
      my.containerName = containerName;
      my.appId = appId;
      my.createTime = new Date().toISOString();
      my.updateTime = new Date().toISOString();
      my.state = 'Running';
      my.appName = app.appName;
      my.devEnv = devEnv;
      this.myContainers.push(my);
    }
  }

  async unbindContainer({
    containerId,
    appId,
    podSelector,
    containerName,
  }: {
    containerId: number;
    appId: number;
    podSelector: string;
    containerName: string;
  }): Promise<any> {
    const index = this.myContainers.findIndex(
      (container) => container.id == containerId,
    );
    if (index < 0) {
      throw new Error(`container ${containerId} not found`);
    }

    this.myContainers[index].podSelector = '';
    this.myContainers[index].appId = '';
    this.myContainers[index].appName = '';
    //this.myContainers[index].containerName = '';
  }
}

export interface ApplicationInfo {
  id: number;
  appName: string;
  devEnv: string;
  createTime: string; //"2023-12-11 09:00:00"
  updateTime: string;

  chart: string;
  entrance: string;
  ide: string;
}

export interface OpenApplicationConfig {
  appid: string;
  path: string;
}

export interface CreateApplicationConfig {
  name: string;
  type: string;

  systemDB: boolean;
  redis: boolean;
  mongodb: boolean;
  postgreSQL: boolean;

  systemCall: boolean;

  img: string;
  ports: string[];

  ingressRouter: boolean;
  traefik: boolean;
  websitePort: string;

  appData: boolean;
  appCache: boolean;
  userData: string[];

  needGpu: boolean;
  requiredGpu: string | null;
  requiredMemory: string;
  osVersion: string;
}

export enum ContainerState {
  Running,
  Waiting,
  Terminated,
}

export interface AppContainer {
  image: string;
  podSelector: string;
  containerName: string;
}

export interface Container {
  image?: string;
  appId: number;
  id: number;
  devEnv: string;
  createTime: string;
  updateTime: string;
  podSelector: string;
  containerName: string;
  appName: string;
  state: ContainerState;
  devContainerName: string;
  icon: string;
}

export enum authLevel {
  private,
  public,
}

export enum OpenMethod {
  default,
  window,
  iframe,
}

export interface AppCfg {
  version: string;
  metadata: {
    name: string;
    icon: string;
    description: string;
    appid: string;
    title: string;
    version: string;
    categories: string[];
    target: string;
  };
  entrances: [
    {
      name: string;
      host: string;
      port: number;
      icon: string;
      title: string;
      authLevel: authLevel;
      openMethod: OpenMethod;
    }
  ];
  spec: {
    versionName: string;
    fullDescription: string;
    upgradeDescription: string;
    promoteImage: string[];
    promoteVideo: string;
    subCategory: string;
    developer: string;
    requiredGpu: string;
    requiredCpu: string;
    requiredMemory: string;
    requiredDisk: string;
    supportClient: {
      edge: string;
      android: string;
      ios: string;
      windows: string;
      mac: string;
      linux: string;
    };
  };
  permission: {
    appData: boolean;
    sysData: [
      {
        group: string;
        dataType: string;
        version: string;
        ops: string[];
      }
    ];
  };
  middleware: {
    postgres: {
      username: string;
      password: string;
      databases: [
        {
          name: string;
          distributed: boolean;
        }
      ];
    };
    redis: {
      password: string;
      databases: [
        {
          name: string;
        }
      ];
    };
    mongodb: {
      username: string;
      password: string;
      databases: [
        {
          name: string;
        }
      ];
    };
    zincSearch: {
      username: string;
      password: string;
      indexes: [
        {
          name: string;
        }
      ];
    };
  };
  options: {
    policies: [
      {
        entranceName: string;
        description: string;
        uriRegex: string;
        level: string;
        oneTime: boolean;
        validDuration: string;
      }
    ];
    analytics: {
      enabled: boolean;
    };
    dependencies: [
      {
        name: string;
        version: string;
        type: string;
      }
    ];
    appScope: {
      clusterScoped: boolean;
      appRef: string[];
    };
    wsConfig: {
      port: number;
      url: string;
    };
  };
}

export interface MenuLabelType {
  label: string;
  key: string;
  icon: string;
  children?: MenuLabelType[];
}

export interface DocumenuType {
  id: string | number;
  message: string;
  link: string;
}

export interface MenuTab {
  name: string;
  status: MenuTabStatus;
}

export enum MenuTabStatus {
  default,
  success,
  error,
}

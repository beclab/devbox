import { Controller, Get, Post, Body, Param, Put, Query } from '@nestjs/common';

import { Result, returnSucceed } from '@bytetrade/core';
//import { ApplicationInfo } from '@devbox/core';
import { AppService } from './app.service';
@Controller('/api')
export class AppController {
  //private readonly logger = new Logger(AppController.name);
  constructor(private readonly appService: AppService) {
    //
  }

  @Post('/command/create-app')
  async createApp(@Body() body: any): Promise<Result<any>> {
    const app = await this.appService.saveApp(body);

    return returnSucceed(app);
  }

  @Post('/command/delete-app')
  async deleteApp(@Body() { name }: { name: string }): Promise<Result<any>> {
    const app = await this.appService.removeApp(name);

    return returnSucceed(app);
  }

  @Get('/command/list-app')
  async listApp(): Promise<Result<any>> {
    return returnSucceed(this.appService.apps);
  }

  @Get('/app-cfg')
  async getCfg(@Query('app') appName: string): Promise<Result<any>> {
    console.log(appName);
    return returnSucceed(await this.appService.getCfg(appName));
  }

  @Put('/app-cfg')
  async putCfg(
    @Query('app') appName: string,
    @Body() body,
  ): Promise<Result<any>> {
    console.log(appName);
    console.log(body);
    await this.appService.setCfg(appName, body);
    return returnSucceed(null);
  }

  @Get('/list-app-containers')
  async listAppContainers(@Query('app') appName: string): Promise<Result<any>> {
    return returnSucceed(await this.appService.getAppContainers(appName));
  }

  @Get('/list-my-containers')
  async listMyContainers(): Promise<Result<any>> {
    return returnSucceed(await this.appService.getMyContainers());
  }

  @Post('/bind-container')
  async bindContainer(@Body() body): Promise<Result<any>> {
    console.log('bind');
    console.log(body);
    return returnSucceed(await this.appService.bindContainer(body));
  }

  @Post('/unbind-container')
  async unbindContainer(@Body() body): Promise<Result<any>> {
    return returnSucceed(await this.appService.unbindContainer(body));
  }
}

# 获取运行信息

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /public/statistics:
    get:
      summary: 获取运行信息
      deprecated: false
      description: 获取 MCSL-Sync 信息
      tags:
        - 公共信息
      parameters: []
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: object
                    properties:
                      name:
                        type: string
                        title: 项目名
                      author:
                        type: string
                        title: 作者
                      version:
                        type: string
                        title: 版本
                      config:
                        type: object
                        properties:
                          url:
                            type: string
                          port:
                            type: integer
                          ssl_cert_path:
                            type: string
                          ssl_key_path:
                            type: string
                          node_list:
                            type: array
                            items:
                              type: string
                        required:
                          - url
                          - port
                          - ssl_cert_path
                          - ssl_key_path
                          - node_list
                        x-apifox-orders:
                          - url
                          - port
                          - ssl_cert_path
                          - ssl_key_path
                          - node_list
                        title: 程序设置
                    required:
                      - name
                      - author
                      - version
                      - config
                    x-apifox-orders:
                      - name
                      - author
                      - version
                      - config
                    title: 数据
                  code:
                    type: integer
                    title: 状态码
                  msg:
                    type: string
                    title: 提示
                required:
                  - data
                  - code
                  - msg
                x-apifox-orders:
                  - data
                  - code
                  - msg
                title: 返回值
              example:
                data:
                  name: MCSL-Sync
                  author: MCSLTeam
                  version: v0.1.0
                  config:
                    url: 0.0.0.0
                    port: 4523
                    ssl_cert_path: ''
                    ssl_key_path: ''
                    node_list:
                      - https://deyang.node.sync.mcsl.com.cn:4523/
                code: 200
                msg: Success!
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 公共信息
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/4093257/apis/api-151671324-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: https://sync.mcsl.com.cn/api
    description: 正式环境
security: []

```
# 获取核心列表

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /core:
    get:
      summary: 获取核心列表
      deprecated: false
      description: 获取 MCSL-Sync 支持的所有核心类型
      tags:
        - 核心获取
      parameters: []
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      type: string
                    title: 支持的核心类型列表
                  code:
                    type: integer
                    title: 状态码
                  msg:
                    type: string
                    title: 提示
                required:
                  - data
                  - code
                  - msg
                x-apifox-orders:
                  - data
                  - code
                  - msg
                title: 返回值
              example:
                data:
                  - Arclight
                  - Lightfall
                  - LightfallClient
                  - Banner
                  - Mohist
                  - Spigot
                  - BungeeCord
                  - Leaves
                  - Pufferfish
                  - Pufferfish+
                  - Pufferfish+Purpur
                  - SpongeForge
                  - SpongeVanilla
                  - Paper
                  - Folia
                  - Travertine
                  - Velocity
                  - Waterfall
                  - Purpur
                  - Purformance
                  - CatServer
                  - Craftbukkit
                  - Vanilla
                  - Fabric
                  - Forge
                code: 200
                msg: Success!
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 核心获取
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/4093257/apis/api-175593126-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: https://sync.mcsl.com.cn/api
    description: 正式环境
security: []

```
# 获取特定核心支持的 Minecraft 版本列表

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /core/{core_type}:
    get:
      summary: 获取特定核心支持的 Minecraft 版本列表
      deprecated: false
      description: ''
      tags:
        - 核心获取
      parameters:
        - name: core_type
          in: path
          description: 核心类型
          required: true
          example: Paper
          schema:
            type: string
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: object
                    properties:
                      type:
                        type: string
                        title: 数据库类别
                      versions:
                        type: array
                        items:
                          type: string
                        title: 支持的 Minecraft 版本
                    required:
                      - type
                      - versions
                    x-apifox-orders:
                      - type
                      - versions
                    title: 该类核心的数据
                  code:
                    type: integer
                    title: 状态码
                  msg:
                    type: string
                    title: 提示
                required:
                  - data
                  - code
                  - msg
                x-apifox-orders:
                  - data
                  - code
                  - msg
                title: 返回值
              examples:
                '1':
                  summary: 成功示例
                  value:
                    data:
                      type: runtime
                      versions:
                        - 1.9.4
                        - 1.8.8
                        - 1.20.6
                        - 1.20.5
                        - 1.20.4
                        - 1.20.2
                        - 1.20.1
                        - '1.20'
                        - 1.19.4
                        - 1.19.3
                        - 1.19.2
                        - 1.19.1
                        - '1.19'
                        - 1.18.2
                        - 1.18.1
                        - '1.18'
                        - 1.17.1
                        - '1.17'
                        - 1.16.5
                        - 1.16.4
                        - 1.16.3
                        - 1.16.2
                        - 1.16.1
                        - 1.15.2
                        - 1.15.1
                        - '1.15'
                        - 1.14.4
                        - 1.14.3
                        - 1.14.2
                        - 1.14.1
                        - '1.14'
                        - 1.13.2
                        - 1.13.1
                        - 1.13-pre7
                        - '1.13'
                        - 1.12.2
                        - 1.12.1
                        - '1.12'
                        - 1.11.2
                        - 1.10.2
                    code: 200
                    msg: Success!
                '2':
                  summary: 成功示例
                  value:
                    data: null
                    code: 404
                    msg: 'Error: No data were found.'
          headers: {}
          x-apifox-name: 成功
        '404':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: 'null'
                    title: 无数据
                  code:
                    type: integer
                    title: 状态码
                  msg:
                    type: string
                    title: 提示
                required:
                  - data
                  - code
                  - msg
                x-apifox-orders:
                  - data
                  - code
                  - msg
          headers: {}
          x-apifox-name: 记录不存在
      security: []
      x-apifox-folder: 核心获取
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/4093257/apis/api-175593188-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: https://sync.mcsl.com.cn/api
    description: 正式环境
security: []

```
# 获取特定核心的特定 Minecraft 版本的构建列表

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /core/{core_type}/{mc_version}:
    get:
      summary: 获取特定核心的特定 Minecraft 版本的构建列表
      deprecated: false
      description: ''
      tags:
        - 核心获取
      parameters:
        - name: core_type
          in: path
          description: 核心类型
          required: true
          example: Paper
          schema:
            type: string
        - name: mc_version
          in: path
          description: Minecraft 版本
          required: true
          example: 1.20.6
          schema:
            type: string
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: object
                    properties:
                      type:
                        type: string
                        title: 数据库类别
                      builds:
                        type: array
                        items:
                          type: string
                        title: 构建列表
                    required:
                      - type
                      - builds
                    x-apifox-orders:
                      - type
                      - builds
                    title: 该 Minecraft 版本的构建信息
                  code:
                    type: integer
                    title: 状态码
                  msg:
                    type: string
                    title: 提示
                required:
                  - data
                  - code
                  - msg
                x-apifox-orders:
                  - data
                  - code
                  - msg
                title: 返回值
              example:
                data:
                  type: runtime
                  builds:
                    - build30
                    - build29
                    - build28
                    - build27
                    - build26
                    - build25
                    - build24
                    - build23
                code: 200
                msg: Success!
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 核心获取
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/4093257/apis/api-175593408-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: https://sync.mcsl.com.cn/api
    description: 正式环境
security: []

```
# 获取特定构建的详细信息

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /core/{core_type}/{mc_version}/{core_version}:
    get:
      summary: 获取特定构建的详细信息
      deprecated: false
      description: ''
      tags:
        - 核心获取
      parameters:
        - name: core_type
          in: path
          description: 核心类型
          required: true
          example: Paper
          schema:
            type: string
        - name: mc_version
          in: path
          description: Minecraft 版本
          required: true
          example: 1.20.6
          schema:
            type: string
        - name: core_version
          in: path
          description: 构建版本
          required: true
          example: build30
          schema:
            type: string
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: object
                    properties:
                      type:
                        type: string
                        title: 数据库类别
                      build:
                        type: object
                        properties:
                          sync_time:
                            type: string
                            title: 官方同步时间
                          download_url:
                            type: string
                            title: 下载地址
                          core_type:
                            type: string
                            title: 核心类型
                          mc_version:
                            type: string
                            title: Minecraft 版本
                          core_version:
                            type: string
                            title: 构建版本
                        required:
                          - sync_time
                          - download_url
                          - core_type
                          - mc_version
                          - core_version
                        x-apifox-orders:
                          - sync_time
                          - download_url
                          - core_type
                          - mc_version
                          - core_version
                        title: 该构建的详细信息
                    required:
                      - type
                      - build
                    x-apifox-orders:
                      - type
                      - build
                  code:
                    type: integer
                    title: 状态码
                  msg:
                    type: string
                    title: 提示
                required:
                  - data
                  - code
                  - msg
                x-apifox-orders:
                  - data
                  - code
                  - msg
                title: 返回值
              example:
                data:
                  type: runtime
                  build:
                    sync_time: '2024-05-01T00:01:24Z'
                    download_url: >-
                      https://deyang.node.sync.mcsl.com.cn:4523/core/Paper/1.20.6/build30/download
                    core_type: Paper
                    mc_version: 1.20.6
                    core_version: build30
                code: 200
                msg: Success!
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 核心获取
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/4093257/apis/api-175593519-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: https://sync.mcsl.com.cn/api
    description: 正式环境
security: []

```

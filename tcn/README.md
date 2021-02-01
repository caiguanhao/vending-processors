# TCN

## Lifter Function Codes

Following function codes can be found in the
[decompiled](http://www.javadecompilers.com/) Java class file in the AAR file
from TCN Lifter Android demo project.

| a/b.java | a/f.java | b/c.java | TcnVendEventID |                                |
|----------|----------|----------|----------------|--------------------------------|
| "01"     | 125      | 251      | 380            | `CMD_QUERY_STATUS_LIFTER`      |
| "02"     | -        | -        | -              | `SHIP`                         |
| "03"     | 127      | 255      | 381            | `CMD_TAKE_GOODS_DOOR`          |
| "04"     | 128      | 257      | 382            | `CMD_LIFTER_UP`                |
| "05"     | 129      | 259      | 383            | `CMD_LIFTER_BACK_HOME`         |
| "06"     | 130      | 261      | 384            | `CMD_CLAPBOARD_SWITCH`         |
| "07"     | 131      | 263      | 385            | `CMD_OPEN_COOL`                |
| "07"     | 132      | 265      | 386            | `CMD_OPEN_HEAT`                |
| "07"     | 133      | 267      | 387            | `CMD_CLOSE_COOL_HEAT`          |
| "50"     | 134      | 269      | 388            | `CMD_CLEAN_FAULTS`             |
| "51"     | 135      | 271      | 389            | `CMD_QUERY_PARAMETERS`         |
| "52"     | ?        | 273 ?    | 390 ?          | `CMD_QUERY_DRIVER_CMD` ?       |
| "53"     | 137      | 275      | 391            | `CMD_SET_SWITCH_OUTPUT_STATUS` |
| "80"     | 138      | 277      | 392            | `CMD_SET_ID`                   |
| "81"     | 139      | 279      | 393            | `CMD_SET_LIGHT_OUT_STEP`       |
| "82"     | 140      | 281      | 394            | `CMD_SET_PARAMETERS`           |
| "83"     | 141      | 283      | 395            | `CMD_FACTORY_RESET`            |
| "84"     | 142      | 285      | 396            | `CMD_DETECT_LIGHT`             |
| "85"     | 143      | 287      | 397            | `CMD_DETECT_SHIP`              |
| "86"     | 144      | 289      | 398 ?          | `CMD_DETECT_SWITCH_INPUT` ?    |

`?` means "not sure".

## Lifter Status Error Codes

| Error Code | "Official" Error Message | English (Google Translate) |
|------------|--------------------------|----------------------------|
| 1   | 锁门时锁开关没检测到位                   | The lock switch is not detected when the door is locked |
| 2   | 锁门时门开关没检测到位                   | The door switch is not detected when the door is locked |
| 3   | 升降电机电流过大                         | Lifting motor current is too large |
| 4   | 超过极限步数还没到底                     | The number of steps beyond the limit has not yet reached the end |
| 5   | 检测到的最大层数比现在要出货的层数还少   | The maximum number of layers detected is less than the number of layers to be shipped now |
| 6   | 回原点运行超时                           | Back to origin running timeout |
| 7   | 正常运行时超时                           | Timeout during normal operation |
| 8   | 下降正常运行时超时                       | Falling timeout during normal operation |
| 9   | 开门时锁开关没检测到位                   | The lock switch is not detected when opening the door |
| 10  | 等待离开层检测光检超时                   | Waiting for the leaving layer to detect the light detection timeout |
| 10i | 升降机光检被挡住                         | The lift light inspection is blocked |
| 20i | 升降机光检不发送也有接收                 | Elevator light inspection does not send and receive |
| 30  | 往上移动了一段距离，但原点开关仍然没放开 | Moved up a distance, but the origin switch is still not released |
| 31  | 推板运行超时                             | Push plate running timeout |
| 32  | 推板电流过大                             | Push plate current is too large |
| 33  | 推板从来没有电流                         | Push plate never has current |
| 34  | 取货口没有货物                           | No goods at the pickup port |
| 35  | 售货前货斗里面有货                       | There is stock in the hopper before sale |
| 36  | 货在货道口被卡住                         | The cargo is stuck at the cargo crossing |
| 37  | 升降电机开路                             | Lift motor open circuit |
| 40  | 货道驱动板故障                           | Cargo Drive Board Failure |
| 41  | FLASH檫除错误                            | FLASH sassafras error |
| 42  | FLASH写错误                              | FLASH write error |
| 43  | 错误命令                                 | Wrong command |
| 44  | 校验错误                                 | Check Error |
| 45  | 柜门没关                                 | The door is not closed |
| 46  | 第二次购买到履带货道                     | The second purchase to the crawler track |
| 47  | 1层超时                                  | Level 1 timeout |
| 48  | 1层过流                                  | 1 layer overcurrent |
| 49  | 1层断线（正反都没有电流）                | 1 layer disconnection (no current on both sides) |
| 50  | 2层超时                                  | Layer 2 timeout |
| 51  | 2层过流                                  | 2 layer overcurrent |
| 52  | 2层断线（正反都没有电流）                | Layer 2 disconnection (no current on both sides) |
| 53  | 3层超时                                  | Layer 3 timeout |
| 54  | 3层过流                                  | 3 layer overcurrent |
| 55  | 3层断线（正反都没有电流）                | 3 layers of disconnection (no current on both sides) |
| 56  | 4层超时                                  | Layer 4 timeout |
| 57  | 4层过流                                  | 4-layer overcurrent |
| 58  | 4层断线（正反都没有电流）                | 4-layer disconnection (no current on both sides) |
| 59  | 5层超时                                  | Level 5 timeout |
| 60  | 5层过流                                  | 5-layer overcurrent |
| 61  | 5层断线（正反都没有电流）                | 5-layer disconnection (no current on both sides) |
| 64  | 无效电机                                 | Invalid motor |
| 80  | 转动超时                                 | Rotation timeout |
| 127 | 驱动板不回复命令                         | The driver board does not respond to commands |

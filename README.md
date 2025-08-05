# System Design

## H-A Stock

固定的股票，这类股票又在港股上市又在A股上市，不需要增删改查，只需要查询股票信息，在前端表格呈现

- 前端表格
  - 股票名称
  - H股票代码
  - H股票价格
  - H股票涨跌幅
  - A股票代码
  - A股票价格
  - A股票涨跌幅
- 需要爬取的股票信息
  - MySQL(d_come的stock_pairs表)
    - 表结构
      - id bigint
      - stock_name varchar(50)(ex:第一拖拉机股份)
      - a_stock_code varchar(50)(ex:601038.SH)
      - h_stock_code varchar(50)(ex:00038.HK)
      - crawl_time datetime(3)
      - updated_at datetime(3)
- 港股数据获取（腾讯股票）
  - 输入
    - 股票代码（ex:hk00038）
  - 输出(redis)
    - 股票名称
    - 股票代码
    - 股票价格
    - 股票涨跌幅
    - 股票涨跌量
    - 股票总手
    - 股票总金额
- A股数据获取（新浪财经）
  - 输入
    - 股票代码（ex:sh601038）
  - 输出(redis)
    - 股票名称
    - 股票代码
    - 股票价格
    - 股票涨跌幅
    - 股票涨跌量
    - 股票总手
    - 股票总金额

## Custom stock

用户自定义的股票，商品，等，同时前端需要增删改查自定义的信息

- 前端表格
  - 名称
  - 代码
  - 价格
  - 涨跌幅
  - 公式计算结果（来自功能三的数据计算）
- 自定义的信息
  - MySQL（d_come的custom表）
  - 表结构
    - id bigint
    - custom_name varchar(50)(ex:第一拖拉机股份)
    - custom_code varchar(50)(ex:601038.SH)
    - crawl_time datetime(3)
    - updated_at datetime(3)
- 数据获取（随机新浪或者腾讯）
  - 输入
    - 代码（ex:sh601038）
  - 输出(redis)
    - 股票名称
    - 股票代码
    - 股票价格
    - 股票涨跌幅
    - 股票涨跌量
    - 股票总手
    - 股票总金额

## 数据计算

自定义编辑latex公式，选择爬取的数据做计算，最后输出结果，需要能增删改查自定义的公式

- latex公式存储
  - MySQL（d_come的latex表）
  - 表结构
    - id bigint
    - latex_name varchar(50)(ex:回归计算)
    - latex_formula varchar(50)(ex:\frac{1}{2})
    - crawl_time datetime(3)
    - updated_at datetime(3)

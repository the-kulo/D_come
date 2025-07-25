import pymysql
import yaml
import pandas as pd
from contextlib import closing
import requests
from datetime import datetime
import time
import random
from dbutils.pooled_db import PooledDB
import json

class StockDataFetcher:
    def __init__(self):
        with open('config.yaml','r',encoding='utf-8') as file:
            self.config = yaml.safe_load(file)
        
        self.DB_CONFIG = {
            'host': self.config['database']['mysql']['host'],
            'port': self.config['database']['mysql']['port'],
            'user': self.config['database']['mysql']['username'],
            'password': self.config['database']['mysql']['password'],
            'database': self.config['database']['mysql']['database'],
            'charset': 'utf8mb4'
        }
        
        try:
            self.pool = PooledDB(
                creator=pymysql,
                maxconnections=20,  
                mincached=2,        
                maxcached=5,       
                maxshared=0,        
                blocking=True,    
                maxusage=20,        
                ping=1,             
                **self.DB_CONFIG,
                autocommit=True
            )
        except Exception as e:
            print(f"数据库连接错误: {e}")
            self.pool = None

    def query_stock_pairs(self):
        """查询股票对数据"""
        sql = "SELECT stock_name, a_stock_code, h_stock_code FROM stock_pairs"
        with closing(self.pool.connection()) as conn:
            with closing(conn.cursor()) as cur:
                cur.execute(sql)
                results = cur.fetchall()
        df = pd.DataFrame(results)
        df.columns = ['股票名称', 'A股股票代码', '港股股票代码']
        return df

    def get_random_user_agent(self):
        """从config中随机选择一个User-Agent"""
        user_agents = self.config.get('crawler', {}).get('user_agent', [])
        if user_agents:
            return random.choice(user_agents)
        return 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'

    def parse_stock_code(self, stock_code):
        """解析数据库中的股票代码格式"""
        if not stock_code or str(stock_code) == 'nan':
            return None
        
        stock_code = str(stock_code).strip()
        
        if '.' in stock_code:
            code_part, exchange_part = stock_code.split('.', 1)
            exchange_part = exchange_part.lower()
            
            if exchange_part == 'sh':
                return f'sh{code_part}'
            elif exchange_part == 'sz':
                return f'sz{code_part}'
            elif exchange_part == 'bj':
                return f'bj{code_part}'
            elif exchange_part in ['hk', 'hkg']:
                return f'hk{code_part.zfill(5)}'
            else:
                print(f"未知的交易所后缀: {exchange_part} (股票代码: {stock_code})")
                return None
        else:
            print(f"股票代码格式异常，缺少交易所后缀: {stock_code}")
            return None

    def get_tencent_hk_price(self, hk_codes):
        """从腾讯财经获取港股价格（使用实时API）"""
        if not hk_codes:
            return {}
        
        # 转换为腾讯实时API需要的格式：r_hk00700
        real_time_codes = []
        for code in hk_codes:
            if code.startswith('hk'):
                # 将 hk00700 转换为 r_hk00700
                real_time_codes.append(f"r_{code}")
            else:
                real_time_codes.append(f"r_hk{code.zfill(5)}")
        
        codes_str = ','.join(real_time_codes)
        url = f"http://qt.gtimg.cn/q={codes_str}"
        
        headers = {
            'User-Agent': self.get_random_user_agent(),
            'Accept': '*/*',
            'Referer': 'http://stockapp.finance.qq.com',
            'Cache-Control': 'no-cache',
            'Pragma': 'no-cache'
        }
        
        try:
            time.sleep(random.uniform(0.1, 0.3))
            
            response = requests.get(url, headers=headers, timeout=10)
            response.encoding = 'gbk'
            
            results = {}
            lines = response.text.strip().split('\n')
            
            for line in lines:
                if '~' in line and line.startswith('v_r_'):
                    # 解析腾讯实时API返回格式：v_r_hk00700="100~腾讯控股~00700~557.000~..."
                    parts = line.split('~')
                    if len(parts) >= 35:
                        try:
                            # 提取股票代码，去掉 v_r_ 前缀
                            code_part = line.split('=')[0].replace('v_r_', '')
                            
                            # 腾讯港股实时API字段映射
                            name = parts[1]                    # 字段1: 股票名称
                            current_price = float(parts[3]) if parts[3] else 0    # 字段3: 当前价格
                            yesterday_close = float(parts[4]) if parts[4] else 0  # 字段4: 昨收价
                            today_open = float(parts[5]) if parts[5] else 0       # 字段5: 今开价
                            today_high = float(parts[33]) if len(parts) > 33 and parts[33] else 0  # 字段33: 最高价
                            today_low = float(parts[34]) if len(parts) > 34 and parts[34] else 0   # 字段34: 最低价
                            volume = int(float(parts[6])) if parts[6] else 0      # 字段6: 成交量
                            
                            # 字段30: 时间字段，格式为 "2025/01/24 16:08:08"
                            datetime_str = parts[30] if len(parts) > 30 and parts[30] else ''
                            
                            # 解析时间格式：2025/01/24 16:08:08
                            if datetime_str and ' ' in datetime_str:
                                try:
                                    date_part, time_part = datetime_str.split(' ')
                                    # 转换日期格式：2025/01/24 -> 2025-01-24
                                    if '/' in date_part:
                                        date_components = date_part.split('/')
                                        if len(date_components) == 3:
                                            date_part = f"{date_components[0]}-{date_components[1].zfill(2)}-{date_components[2].zfill(2)}"
                                    
                                    # 验证时间格式是否正确
                                    if ':' not in time_part or len(time_part.split(':')) != 3:
                                        # 如果时间格式异常，使用当前时间
                                        from datetime import datetime
                                        time_part = datetime.now().strftime('%H:%M:%S')
                                        
                                except (ValueError, IndexError):
                                    # 解析失败，使用当前时间
                                    from datetime import datetime
                                    now = datetime.now()
                                    date_part = now.strftime('%Y-%m-%d')
                                    time_part = now.strftime('%H:%M:%S')
                            else:
                                # 如果时间字段为空，使用当前时间
                                from datetime import datetime
                                now = datetime.now()
                                date_part = now.strftime('%Y-%m-%d')
                                time_part = now.strftime('%H:%M:%S')
                            
                            results[code_part] = {
                                'name': name,
                                'current_price': current_price,
                                'yesterday_close': yesterday_close,
                                'today_open': today_open,
                                'today_high': today_high,
                                'today_low': today_low,
                                'volume': volume,
                                'date': date_part,
                                'time': time_part
                            }
                            
                            # 计算涨跌幅
                            if yesterday_close > 0:
                                change = current_price - yesterday_close
                                change_percent = (change / yesterday_close) * 100
                                results[code_part]['change'] = change
                                results[code_part]['change_percent'] = change_percent
                            else:
                                results[code_part]['change'] = 0
                                results[code_part]['change_percent'] = 0
                                
                        except (ValueError, IndexError) as e:
                            print(f"解析腾讯港股实时数据失败: {e}, 数据: {line}")
                            continue
            
            return results
        except Exception as e:
            print(f"获取腾讯港股实时数据失败: {e}")
            return {}

    def get_mixed_stock_price(self, stock_codes):
        """混合获取股票价格：A股用新浪，港股用腾讯"""
        if not stock_codes:
            return {}
        
        # 分离A股和港股代码
        a_stock_codes = [code for code in stock_codes if not code.startswith('hk')]
        hk_stock_codes = [code for code in stock_codes if code.startswith('hk')]
        
        results = {}
        
        # 获取A股数据（使用新浪API）
        if a_stock_codes:
            a_results = self.get_sina_stock_price(a_stock_codes)
            results.update(a_results)
        
        # 获取港股数据（使用腾讯API）
        if hk_stock_codes:
            hk_results = self.get_tencent_hk_price(hk_stock_codes)
            results.update(hk_results)
        
        return results

    def get_sina_stock_price(self, stock_codes):
        """从新浪财经获取A股价格"""
        if not stock_codes:
            return {}
        
        # 过滤掉港股代码，只处理A股
        a_stock_codes = [code for code in stock_codes if not code.startswith('hk')]
        if not a_stock_codes:
            return {}
        
        codes_str = ','.join(a_stock_codes)
        url = f"https://hq.sinajs.cn/list={codes_str}"
        
        headers = {
            'User-Agent': self.get_random_user_agent(),
            'Accept': '*/*',
            'Accept-Encoding': 'gzip, deflate, br',
            'Referer': 'https://finance.sina.com.cn',
            'Content-Type': 'application/javascript; charset=GB18030',
            'Cache-Control': 'no-cache',
            'Pragma': 'no-cache'
        }
        
        try:
            time.sleep(random.uniform(0.1, 0.3))
            
            response = requests.get(url, headers=headers, timeout=10)
            response.encoding = 'gbk'
            
            results = {}
            lines = response.text.strip().split('\n')
            
            for line in lines:
                if '=' in line and '"' in line:
                    code = line.split('=')[0].replace('var hq_str_', '')
                    data_str = line.split('"')[1]
                    
                    if data_str:
                        data_parts = data_str.split(',')
                        
                        # A股数据格式处理
                        if len(data_parts) >= 32:
                            try:
                                current_price = float(data_parts[3]) if data_parts[3] else 0
                                yesterday_close = float(data_parts[2]) if data_parts[2] else 0
                                
                                results[code] = {
                                    'name': data_parts[0],
                                    'current_price': current_price,
                                    'yesterday_close': yesterday_close,
                                    'today_open': float(data_parts[1]) if data_parts[1] else 0,
                                    'today_high': float(data_parts[4]) if data_parts[4] else 0,
                                    'today_low': float(data_parts[5]) if data_parts[5] else 0,
                                    'volume': int(data_parts[8]) if data_parts[8] else 0,
                                    'date': data_parts[30],
                                    'time': data_parts[31]
                                }
                                
                                if yesterday_close > 0:
                                    change = current_price - yesterday_close
                                    change_percent = (change / yesterday_close) * 100
                                    results[code]['change'] = change
                                    results[code]['change_percent'] = change_percent
                                else:
                                    results[code]['change'] = 0
                                    results[code]['change_percent'] = 0
                                    
                            except (ValueError, IndexError):
                                continue
            
            return results
        except Exception as e:
            print(f"获取新浪A股数据失败: {e}")
            return {}

    def get_all_stock_data(self):
        """获取所有股票数据"""
        try:
            # 获取股票对数据
            df = self.query_stock_pairs()
            
            # 收集所有股票代码
            all_codes = []
            code_mapping = {}
            
            for _, row in df.iterrows():
                # 处理A股代码
                if row['A股股票代码'] and str(row['A股股票代码']) != 'nan':
                    parsed_a = self.parse_stock_code(row['A股股票代码'])
                    if parsed_a:
                        all_codes.append(parsed_a)
                        code_mapping[parsed_a] = row['股票名称']
                
                # 处理H股代码
                if row['港股股票代码'] and str(row['港股股票代码']) != 'nan':
                    parsed_h = self.parse_stock_code(row['港股股票代码'])
                    if parsed_h:
                        all_codes.append(parsed_h)
                        code_mapping[parsed_h] = row['股票名称']
            
            # 使用混合API获取价格数据
            prices = self.get_mixed_stock_price(all_codes)
            
            # 构建完整数据
            result = {
                'timestamp': datetime.now().isoformat(),
                'stock_pairs': df.to_dict('records'),
                'prices': prices,
                'code_mapping': code_mapping,
                'success_count': len([p for p in prices.values() if p['current_price'] > 0]),
                'total_count': len(df)
            }
            
            return result
            
        except Exception as e:
            print(f"获取股票数据时发生错误: {e}")
            return {
                'timestamp': datetime.now().isoformat(),
                'error': str(e),
                'stock_pairs': [],
                'prices': {},
                'code_mapping': {},
                'success_count': 0,
                'total_count': 0
            }

    def save_data_to_file(self, data, filename='stock_data.json'):
        """将数据保存到文件"""
        try:
            with open(filename, 'w', encoding='utf-8') as f:
                json.dump(data, f, ensure_ascii=False, indent=2)
            return True
        except Exception as e:
            print(f"保存数据失败: {e}")
            return False

if __name__ == "__main__":
    fetcher = StockDataFetcher()
    data = fetcher.get_all_stock_data()
    fetcher.save_data_to_file(data)
    print(f"数据获取完成，成功获取 {data['success_count']} 只股票数据")
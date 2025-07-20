import requests
import time
import random
import yaml
import json

# 读取配置文件
with open('config.yaml','r',encoding='utf-8') as file:
    config = yaml.safe_load(file)

def get_random_user_agent():
    """从config中随机选择一个User-Agent"""
    user_agents = config.get('crawler', {}).get('user_agent', [])
    if user_agents:
        return random.choice(user_agents)
    return 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'

def test_sina_stock_api():
    """测试新浪财经API，打印JSON格式数据"""
    
    # 测试几个股票代码
    test_codes = [
        'sh600036',  # 招商银行A股
        'hk00939'    # 建设银行H股
    ]
    
    codes_str = ','.join(test_codes)
    url = f"https://hq.sinajs.cn/list={codes_str}"
    
    headers = {
        'User-Agent': get_random_user_agent(),
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
                    
                    # 判断是否为港股
                    if code.startswith('hk'):
                        # 港股数据格式
                        if len(data_parts) >= 7:
                            current_price = float(data_parts[6]) if data_parts[6] else 0
                            yesterday_close = float(data_parts[3]) if data_parts[3] else 0
                            
                            results[code] = {
                                'name': data_parts[1],
                                'current_price': current_price,
                                'yesterday_close': yesterday_close,
                                'today_open': float(data_parts[2]) if data_parts[2] else 0,
                                'today_high': float(data_parts[4]) if data_parts[4] else 0,
                                'today_low': float(data_parts[5]) if data_parts[5] else 0,
                                'volume': int(float(data_parts[12])) if len(data_parts) > 12 and data_parts[12] else 0,
                                'date': data_parts[17] if len(data_parts) > 17 else '',
                                'time': data_parts[18] if len(data_parts) > 18 else '',
                                'raw_data_length': len(data_parts),
                                'stock_type': 'H股'
                            }
                            
                            # 计算涨跌幅
                            if yesterday_close > 0:
                                change = current_price - yesterday_close
                                change_percent = (change / yesterday_close) * 100
                                results[code]['change'] = change
                                results[code]['change_percent'] = change_percent
                    else:
                        # A股数据格式
                        if len(data_parts) >= 32:
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
                                'time': data_parts[31],
                                'raw_data_length': len(data_parts),
                                'stock_type': 'A股'
                            }
                            
                            # 计算涨跌幅
                            if yesterday_close > 0:
                                change = current_price - yesterday_close
                                change_percent = (change / yesterday_close) * 100
                                results[code]['change'] = change
                                results[code]['change_percent'] = change_percent
        
        # 打印JSON格式数据
        print(json.dumps(results, ensure_ascii=False, indent=2))
        
    except Exception as e:
        print(f"请求失败: {e}")

if __name__ == "__main__":
    test_sina_stock_api()
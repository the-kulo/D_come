import time
import schedule
from stock_data import StockDataFetcher
from datetime import datetime
import threading
import json

class StockDataScheduler:
    def __init__(self):
        self.fetcher = StockDataFetcher()
        self.last_update = None
        self.is_running = False
        
    def fetch_and_save_data(self):
        """获取并保存股票数据"""
        try:
            print(f"[{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}] 开始获取股票数据...")
            
            data = self.fetcher.get_all_stock_data()
            success = self.fetcher.save_data_to_file(data, 'stock_data.json')
            
            if success:
                self.last_update = datetime.now()
                print(f"[{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}] 数据更新成功，获取 {data['success_count']} 只股票")
            else:
                print(f"[{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}] 数据保存失败")
                
        except Exception as e:
            print(f"[{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}] 获取数据时发生错误: {e}")
    
    def start_scheduler(self):
        """启动定时任务"""
        # 立即执行一次
        self.fetch_and_save_data()
        
        # 设置每10秒执行一次
        schedule.every(10).seconds.do(self.fetch_and_save_data)
        
        self.is_running = True
        print("股票数据定时任务已启动，每10秒更新一次...")
        
        while self.is_running:
            schedule.run_pending()
            time.sleep(1)
    
    def stop_scheduler(self):
        """停止定时任务"""
        self.is_running = False
        print("股票数据定时任务已停止")
    
    def start_in_background(self):
        """在后台线程中启动定时任务"""
        thread = threading.Thread(target=self.start_scheduler, daemon=True)
        thread.start()
        return thread

if __name__ == "__main__":
    scheduler = StockDataScheduler()
    try:
        scheduler.start_scheduler()
    except KeyboardInterrupt:
        scheduler.stop_scheduler()
        print("程序已退出")
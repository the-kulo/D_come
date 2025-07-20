import streamlit as st
from streamlit_autorefresh import st_autorefresh
from dbutils.pooled_db import PooledDB  
import pymysql
import yaml
import pandas as pd
from contextlib import closing
import requests
import json
import os
from datetime import datetime, timedelta
import time

st.set_page_config(
    page_title="AH股票数据分析",
    page_icon="📈",
    layout="wide",  
    initial_sidebar_state="collapsed"  
)

# 每隔 10 秒自动刷新页面
count = st_autorefresh(interval=10 * 1000, key="auto_refresh")

st.title('AH股票数据分析')

# 加载CSS样式
def load_css():
    try:
        with open('styles.css', 'r', encoding='utf-8') as f:
            css_content = f.read()
        st.markdown(f'<style>{css_content}</style>', unsafe_allow_html=True)
    except FileNotFoundError:
        st.warning("CSS样式文件未找到，使用默认样式")

load_css()

# 初始化session state
if 'last_refresh_time' not in st.session_state:
    st.session_state.last_refresh_time = None
if 'last_refresh_success' not in st.session_state:
    st.session_state.last_refresh_success = None

def load_stock_data():
    """从文件加载股票数据"""
    try:
        if os.path.exists('stock_data.json'):
            with open('stock_data.json', 'r', encoding='utf-8') as f:
                data = json.load(f)
            return data
        else:
            return None
    except Exception as e:
        st.error(f"加载数据失败: {e}")
        return None

def highlight_global_search(s, search_term):
    """全局搜索高亮函数"""
    if not search_term:
        return [''] * len(s)
    
    return [
        'background-color: yellow; font-weight: bold'
        if pd.notna(val) and search_term.lower() in str(val).lower()
        else ''
        for val in s
    ]

def apply_color_styling(styled_df):
    """为涨跌幅列应用颜色样式"""
    def color_change_percent(val):
        if pd.isna(val) or val == "0.00%" or val == "--":
            return 'color: gray'
        elif '+' in str(val):
            return 'color: red'  # 上涨用红色
        elif '-' in str(val):
            return 'color: green'  # 下跌用绿色
        else:
            return 'color: gray'
    
    # 为A股和H股涨跌幅列应用颜色
    styled_df = styled_df.applymap(color_change_percent, subset=['A股涨跌幅'])
    styled_df = styled_df.applymap(color_change_percent, subset=['H股涨跌幅'])
    
    return styled_df

def create_display_dataframe(stock_data):
    """根据股票数据创建显示用的DataFrame"""
    if not stock_data or 'stock_pairs' not in stock_data:
        return pd.DataFrame()
    
    df = pd.DataFrame(stock_data['stock_pairs'])
    prices = stock_data.get('prices', {})
    
    # 添加价格列
    df['A股价格'] = ''
    df['A股涨跌幅'] = ''
    df['A股更新时间'] = ''
    df['H股价格'] = ''
    df['H股涨跌幅'] = ''
    df['H股更新时间'] = ''
    
    for idx, row in df.iterrows():
        # 处理A股数据
        a_code = row.get('A股股票代码')
        if a_code and str(a_code) != 'nan':
            # 解析A股代码
            if '.' in str(a_code):
                code_part, exchange_part = str(a_code).split('.', 1)
                if exchange_part.lower() == 'sh':
                    sina_code = f'sh{code_part}'
                elif exchange_part.lower() == 'sz':
                    sina_code = f'sz{code_part}'
                elif exchange_part.lower() == 'bj':
                    sina_code = f'bj{code_part}'
                else:
                    sina_code = None
                
                if sina_code and sina_code in prices:
                    price_info = prices[sina_code]
                    if price_info['current_price'] > 0:
                        df.at[idx, 'A股价格'] = f"¥{price_info['current_price']:.2f}"
                        df.at[idx, 'A股涨跌幅'] = f"{price_info['change_percent']:+.2f}%"
                        df.at[idx, 'A股更新时间'] = f"{price_info['time']}"
                    else:
                        df.at[idx, 'A股价格'] = "停牌"
                        df.at[idx, 'A股涨跌幅'] = "0.00%"
                        df.at[idx, 'A股更新时间'] = "--:--:--"
        
        # 处理H股数据
        h_code = row.get('港股股票代码')
        if h_code and str(h_code) != 'nan':
            # 解析H股代码
            if '.' in str(h_code):
                code_part, exchange_part = str(h_code).split('.', 1)
                if exchange_part.lower() in ['hk', 'hkg']:
                    sina_code = f'hk{code_part.zfill(5)}'
                else:
                    sina_code = None
                
                if sina_code and sina_code in prices:
                    price_info = prices[sina_code]
                    if price_info['current_price'] > 0:
                        df.at[idx, 'H股价格'] = f"HK${price_info['current_price']:.2f}"
                        df.at[idx, 'H股涨跌幅'] = f"{price_info['change_percent']:+.2f}%"
                        df.at[idx, 'H股更新时间'] = f"{price_info['time']}"
                    else:
                        df.at[idx, 'H股价格'] = "停牌"
                        df.at[idx, 'H股涨跌幅'] = "0.00%"
                        df.at[idx, 'H股更新时间'] = "--:--:--"
    
    return df

# 加载股票数据
stock_data = load_stock_data()

# 显示刷新状态
current_time = datetime.now()

if stock_data:
    data_time = datetime.fromisoformat(stock_data['timestamp'])
    time_diff = (current_time - data_time).total_seconds()
    
    refresh_status = f"🕐 **现在时间**: {current_time.strftime('%Y-%m-%d %H:%M:%S')}"
    refresh_status += f" | 🔄 **数据时间**: {data_time.strftime('%Y-%m-%d %H:%M:%S')}"
    refresh_status += f" | ⏱️ **数据年龄**: {int(time_diff)}秒"
    refresh_status += f" | 🔄 **页面刷新次数**: {count}"
    
    if time_diff < 15:  # 15秒内的数据认为是新鲜的
        refresh_status += f" | ✅ **状态**: 数据新鲜"
    else:
        refresh_status += f" | ⚠️ **状态**: 数据较旧"
else:
    refresh_status = f"🕐 **现在时间**: {current_time.strftime('%Y-%m-%d %H:%M:%S')} | 🔄 **页面刷新次数**: {count} | ❌ **状态**: 无数据"

st.info(refresh_status)

# 显示结果
if stock_data and stock_data.get('stock_pairs'):
    # 创建显示用的DataFrame
    display_df = create_display_dataframe(stock_data)
    
    if not display_df.empty:
        # 显示统计信息
        success_count = stock_data.get('success_count', 0)
        total_count = stock_data.get('total_count', 0)
        st.success(f"✅ 成功获取 {success_count} 只股票的价格信息 | 共 {total_count} 只股票")
        
        # 搜索框容器 - 放在成功信息下面，靠右对齐
        st.markdown('<div class="search-container">', unsafe_allow_html=True)
        col1, col2 = st.columns([3, 1])
        with col2:
            search_term = st.text_input('', placeholder="🔍 输入关键词搜索...", label_visibility="collapsed")
        st.markdown('</div>', unsafe_allow_html=True)
        
        # 如果有搜索条件，过滤数据
        if search_term:
            mask = (
                display_df['股票名称'].str.contains(search_term, case=False, na=False) |
                display_df['A股股票代码'].str.contains(search_term, case=False, na=False) |
                display_df['港股股票代码'].str.contains(search_term, case=False, na=False)
            )
            filtered_df = display_df[mask]
            
            if not filtered_df.empty:
                styled_df = filtered_df.style
                styled_df = styled_df.apply(lambda x: highlight_global_search(x, search_term), axis=0)
                styled_df = apply_color_styling(styled_df)
                st.table(styled_df)
            else:
                st.warning(f'未找到包含 "{search_term}" 的股票')
        else:
            # 显示所有数据
            styled_df = display_df.style
            styled_df = apply_color_styling(styled_df)
            st.table(styled_df)
    else:
        st.warning('数据处理失败')
else:
    st.warning('暂无股票数据，请确保数据获取服务正在运行')

import streamlit as st
from dbutils.pooled_db import PooledDB  
import pymysql
import yaml
import pandas as pd
from contextlib import closing

st.title('A-H股票数据分析')

with open('config.yaml','r',encoding='utf-8') as file:
    config = yaml.safe_load(file)

DB_CONFIG = {
    'host': config['database']['mysql']['host'],
    'port': config['database']['mysql']['port'],
    'user': config['database']['mysql']['username'],
    'password': config['database']['mysql']['password'],
    'database': config['database']['mysql']['database'],
    'charset': 'utf8mb4'
}

try:
    pool = PooledDB(
        creator=pymysql,
        maxconnections=20,  
        mincached=2,        
        maxcached=5,       
        maxshared=0,        
        blocking=True,    
        maxusage=20,        
        ping=1,             
        **DB_CONFIG,
        autocommit=True
    )
    
    # test
    with closing(pool.connection()) as conn:
        with closing(conn.cursor()) as cursor:
            cursor.execute("SELECT 1")
            result = cursor.fetchone()
            print(f"database test result: {result}")
            
except Exception as e:
    print(f"database connection error: {e}")
    pool = None

def query_stock_pairs():
    sql = "SELECT stock_name, a_stock_code, h_stock_code FROM stock_pairs"
    with closing(pool.connection()) as conn:
        with closing(conn.cursor()) as cur:
            cur.execute(sql)
            results = cur.fetchall()
    df = pd.DataFrame(results)
    df.columns = ['股票名称', 'A股股票代码', '港股股票代码']
    return df

df = query_stock_pairs()

col1,col2,col3 = st.columns(3)
with col1:
    stock_name = st.text_input('请输入股票名称')
with col2:
    a_stock_code = st.text_input('请输入A股股票代码')
with col3:
    h_stock_code = st.text_input('请输入港股股票代码')

filtered_df = df.copy()

if stock_name:
    filtered_df = filtered_df[filtered_df['股票名称'].str.contains(stock_name, case=False, na=False)]
if a_stock_code:
    filtered_df = filtered_df[filtered_df['A股股票代码'].str.contains(a_stock_code, case=False, na=False)]
if h_stock_code:
    filtered_df = filtered_df[filtered_df['港股股票代码'].str.contains(h_stock_code, case=False, na=False)]

def highlight_specific_column(s, search_term, column_name):
    if not search_term or s.name != column_name:
        return [''] * len(s)
    
    return [
        'background-color: yellow; font-weight: bold'
        if pd.notna(val) and search_term.lower() in str(val).lower()
        else ''
        for val in s
    ]

if not filtered_df.empty:
    styled_df = filtered_df.style
    
    if stock_name:
        styled_df = styled_df.apply(lambda x: highlight_specific_column(x, stock_name, '股票名称'), axis=0)
    if a_stock_code:
        styled_df = styled_df.apply(lambda x: highlight_specific_column(x, a_stock_code, 'A股股票代码'), axis=0)
    if h_stock_code:
        styled_df = styled_df.apply(lambda x: highlight_specific_column(x, h_stock_code, '港股股票代码'), axis=0)
    
    st.table(styled_df)
else:
    st.warning('未查询到相关股票')
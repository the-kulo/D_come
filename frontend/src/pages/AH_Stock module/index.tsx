import React, { useState, useEffect } from 'react'
import './AH_Stock.css'

interface StockData {
    name: string
    code: string
    price: number
    change: number
    changeValue: number
    volume: number
    amount: number
}

interface HAStockData {
    stockName: string
    aStockCode: string
    aStockData?: StockData
    aUpdateTime?: string  // A股更新时间
    hStockCode: string
    hStockData?: StockData
    hUpdateTime?: string  // H股更新时间
}

const AH_Stock: React.FC = () => {
    const [stocks, setStocks] = useState<HAStockData[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [searchTerm, setSearchTerm] = useState('')
    const [allStocks, setAllStocks] = useState<HAStockData[]>([]) // 保存所有数据用于搜索
    const [connectionStatus, setConnectionStatus] = useState<'connected' | 'disconnected' | 'checking'>('checking')
    const [lastUpdateTime, setLastUpdateTime] = useState<string>('') // 最后更新时间

    const API_BASE_URL = '/api'

    useEffect(() => {
        // 初始化加载数据
        fetchRealTimeData()
        
        // 设置定时刷新（默认3秒间隔）
        const intervalId = setInterval(() => {
            fetchRealTimeData() // 实时数据，不使用缓存
        }, 3000)

        return () => {
            clearInterval(intervalId)
        }
    }, [])

    // 获取实时数据（不使用缓存）
    const fetchRealTimeData = async (showLoading = false) => {
        if (showLoading) {
            setLoading(true)
        }
        setError(null)
        setConnectionStatus('checking')
        
        try {
            // 使用实时API端点
            const response = await fetch(`${API_BASE_URL}/stocks/ha-pairs/realtime`, {
                method: 'GET',
                headers: {
                    'Cache-Control': 'no-cache',
                    'Pragma': 'no-cache'
                }
            })
            
            if (!response.ok) {
                throw new Error(`服务器响应错误: ${response.status} ${response.statusText}`)
            }
            
            const data: HAStockData[] = await response.json()
            
            if (!Array.isArray(data)) {
                throw new Error('服务器返回的数据格式不正确')
            }
            
            setStocks(data)
            setAllStocks(data)
            setConnectionStatus('connected')
            setLastUpdateTime(new Date().toLocaleTimeString()) // 更新最后刷新时间
            
            if (data.length === 0) {
                setError('数据库中暂无H-A股票数据，请联系管理员添加股票数据')
            }
            
        } catch (error) {
            console.error('获取实时股票数据失败:', error)
            setConnectionStatus('disconnected')
            
            if (error instanceof TypeError && error.message.includes('fetch')) {
                setError('无法连接到后端服务器，请确保服务器正在运行 (http://localhost:8080/api)')
            } else {
                setError(`获取实时数据失败: ${error instanceof Error ? error.message : '未知错误'}`)
            }
            
            // 清空数据，不使用模拟数据
            setStocks([])
            setAllStocks([])
        } finally {
            if (showLoading) {
                setLoading(false)
            }
        }
    }

    const handleRefresh = async () => {
        await fetchRealTimeData(true)
    }

    const handleSearch = () => {
        if (!searchTerm.trim()) {
            setStocks(allStocks)
            return
        }
        
        const filtered = allStocks.filter(stock => 
            stock.stockName.includes(searchTerm) ||
            stock.aStockCode.includes(searchTerm) ||
            stock.hStockCode.includes(searchTerm)
        )
        setStocks(filtered)
    }

    const formatChangeRate = (rate?: number) => {
        if (rate === undefined || rate === null) return '--'
        const sign = rate >= 0 ? '+' : ''
        const className = rate >= 0 ? 'positive' : 'negative'
        return <span className={className}>{sign}{rate.toFixed(2)}%</span>
    }

    const formatPrice = (price?: number) => {
        return price ? price.toFixed(2) : '--'
    }

    const formatVolume = (volume?: number) => {
        if (!volume) return '--'
        if (volume >= 10000) {
            return (volume / 10000).toFixed(1) + '万'
        }
        return volume.toString()
    }

    const formatAmount = (amount?: number) => {
        if (!amount) return '--'
        if (amount >= 100000000) {
            return (amount / 100000000).toFixed(2) + '亿'
        } else if (amount >= 10000) {
            return (amount / 10000).toFixed(1) + '万'
        }
        return amount.toFixed(0)
    }

    const formatUpdateTime = (updateTime?: string) => {
        if (!updateTime) return '--'
        // 只显示时间部分，去掉日期
        const timePart = updateTime.split(' ')[1]
        return timePart || updateTime
    }

    return (
        <div className='ah-stock-page'>
            <h2>H-A股票对比</h2>

            {/* Connection Status */}
            <div className={`connection-status ${connectionStatus}`}>
                <span className='status-indicator'></span>
                {connectionStatus === 'connected' && `已连接到数据库 (最后更新: ${lastUpdateTime})`}
                {connectionStatus === 'disconnected' && '数据库连接失败'}
                {connectionStatus === 'checking' && '检查连接中...'}
            </div>

            {/* Error message */}
            {error && (
                <div className='error-message'>
                    <strong>错误:</strong> {error}
                    {connectionStatus === 'disconnected' && (
                        <div className='error-help'>
                            <p>可能的解决方案:</p>
                            <ul>
                                <li>确保后端服务器正在运行 (端口 8080)</li>
                                <li>检查数据库连接配置</li>
                                <li>确认网络连接正常</li>
                            </ul>
                        </div>
                    )}
                </div>
            )}

            {/* Search section */}
            <div className='search-section'>
                <input 
                    type='text' 
                    placeholder='搜索股票名称或代码' 
                    className='search-input'
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                    disabled={connectionStatus === 'disconnected'}
                />
                <button 
                    className='search-button' 
                    onClick={handleSearch}
                    disabled={connectionStatus === 'disconnected'}
                >
                    搜索
                </button>
                <button 
                    className='refresh-button' 
                    onClick={handleRefresh}
                    disabled={loading || connectionStatus === 'disconnected'}
                >
                    {loading ? '刷新中...' : '立即刷新'}
                </button>
            </div>

            {/* Table section */} 
            <div className='table-section'>
                {loading ? (
                    <div className='loading'>
                        <div className='loading-spinner'></div>
                        <span>正在获取实时数据...</span>
                    </div>
                ) : connectionStatus === 'disconnected' ? (
                    <div className='no-connection'>
                        <h3>无法连接到数据源</h3>
                        <p>请检查后端服务器状态和数据库连接</p>
                        <button onClick={() => fetchRealTimeData()} className='retry-button'>
                            重试连接
                        </button>
                    </div>
                ) : stocks.length === 0 ? (
                    <div className='no-data'>
                        <h3>暂无数据</h3>
                        <p>数据库中没有H-A股票数据，请联系管理员添加数据</p>
                    </div>
                ) : (
                    <table className='stock-table'>
                        <thead>
                            <tr>
                                <th>序号</th>
                                <th>股票名称</th>
                                <th>H股代码</th>
                                <th>H股价格</th>
                                <th>H股涨跌幅</th>
                                <th>H股成交量</th>
                                <th>H股成交额</th>
                                <th>H股更新时间</th>
                                <th>A股代码</th>
                                <th>A股价格</th>
                                <th>A股涨跌幅</th>
                                <th>A股成交量</th>
                                <th>A股成交额</th>
                                <th>A股更新时间</th>
                            </tr>
                        </thead>
                        <tbody>
                            {stocks.map((stock, index) => (
                                <tr key={index}>
                                    <td>{index + 1}</td>
                                    <td className='stock-name'>{stock.stockName}</td>
                                    <td>{stock.hStockCode}</td>
                                    <td>{formatPrice(stock.hStockData?.price)}</td>
                                    <td>{formatChangeRate(stock.hStockData?.change)}</td>
                                    <td>{formatVolume(stock.hStockData?.volume)}</td>
                                    <td>{formatAmount(stock.hStockData?.amount)}</td>
                                    <td className='update-time'>{formatUpdateTime(stock.hUpdateTime)}</td>
                                    <td>{stock.aStockCode}</td>
                                    <td>{formatPrice(stock.aStockData?.price)}</td>
                                    <td>{formatChangeRate(stock.aStockData?.change)}</td>
                                    <td>{formatVolume(stock.aStockData?.volume)}</td>
                                    <td>{formatAmount(stock.aStockData?.amount)}</td>
                                    <td className='update-time'>{formatUpdateTime(stock.aUpdateTime)}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>
        </div>
    )
}

export default AH_Stock
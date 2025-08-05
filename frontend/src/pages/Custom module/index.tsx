import { useState, useEffect } from 'react'
import './Custom.css'

interface CustomStock {
    id?: number
    customName: string
    customCode: string
    price?: number
    changeRate?: number
    changeValue?: number
    volume?: number
    amount?: number
    formulaResult?: number
    updateTime?: string
}

const API_BASE_URL = '/api'

function Custom() {
    const [stocks, setStocks] = useState<CustomStock[]>([])
    const [loading, setLoading] = useState(false)
    const [showAddForm, setShowAddForm] = useState(false)
    const [editingStock, setEditingStock] = useState<CustomStock | null>(null)
    const [formData, setFormData] = useState<CustomStock>({
        customName: '',
        customCode: ''
    })

    useEffect(() => {
        fetchCustomStocks()
    }, [])

    const fetchCustomStocks = async () => {
        setLoading(true)
        try {
            console.log('正在请求:', `${API_BASE_URL}/custom-stocks`)
            const response = await fetch(`${API_BASE_URL}/custom-stocks`)
            console.log('响应状态:', response.status)
            console.log('响应头:', response.headers)
            
            if (response.ok) {
                const result = await response.json()
                console.log('响应数据:', result)
                setStocks(result.data || [])
            } else {
                const errorText = await response.text()
                console.error('获取自定义股票失败:', response.status, errorText)
                alert(`获取数据失败: ${response.status} - ${errorText}`)
            }
        } catch (error) {
            console.error('网络错误:', error)
            alert(`网络错误: ${error}`)
        } finally {
            setLoading(false)
        }
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setLoading(true)
        
        try {
            const url = editingStock 
                ? `${API_BASE_URL}/custom-stocks/${editingStock.id}`
                : `${API_BASE_URL}/custom-stocks`
            
            const method = editingStock ? 'PUT' : 'POST'
            
            const response = await fetch(url, {
                method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(formData)
            })

            if (response.ok) {
                await fetchCustomStocks() // 重新获取数据
                resetForm()
            } else {
                const error = await response.json()
                alert(`操作失败: ${error.error}`)
            }
        } catch (error) {
            console.error('操作失败:', error)
            alert('网络错误，请稍后重试')
        } finally {
            setLoading(false)
        }
    }

    const handleEdit = (stock: CustomStock) => {
        setEditingStock(stock)
        setFormData({
            customName: stock.customName,
            customCode: stock.customCode
        })
        setShowAddForm(true)
    }

    const handleDelete = async (id: number) => {
        if (!window.confirm('确定要删除这个自定义股票吗？')) {
            return
        }

        setLoading(true)
        try {
            const response = await fetch(`${API_BASE_URL}/custom-stocks/${id}`, {
                method: 'DELETE'
            })

            if (response.ok) {
                await fetchCustomStocks() // 重新获取数据
            } else {
                const error = await response.json()
                alert(`删除失败: ${error.error}`)
            }
        } catch (error) {
            console.error('删除失败:', error)
            alert('网络错误，请稍后重试')
        } finally {
            setLoading(false)
        }
    }

    const resetForm = () => {
        setFormData({ customName: '', customCode: '' })
        setEditingStock(null)
        setShowAddForm(false)
    }

    const formatChangeRate = (rate?: number) => {
        if (rate === undefined) return '--'
        const sign = rate >= 0 ? '+' : ''
        const className = rate >= 0 ? 'positive' : 'negative'
        return <span className={className}>{sign}{rate.toFixed(2)}%</span>
    }

    const formatPrice = (price?: number) => {
        return price !== undefined ? price.toFixed(2) : '--'
    }

    const formatVolume = (volume?: number) => {
        if (!volume) return '--'
        if (volume >= 10000) {
            return `${(volume / 10000).toFixed(1)}万`
        }
        return volume.toString()
    }

    const formatAmount = (amount?: number) => {
        if (!amount) return '--'
        if (amount >= 100000000) {
            return `${(amount / 100000000).toFixed(2)}亿`
        } else if (amount >= 10000) {
            return `${(amount / 10000).toFixed(1)}万`
        }
        return amount.toFixed(2)
    }

    return (
        <div className='custom-page'>
            <div className='header'>
                <h2>自定义股票管理</h2>
                <button 
                    className='add-button'
                    onClick={() => setShowAddForm(true)}
                    disabled={loading}
                >
                    添加股票
                </button>
            </div>

            {/* 添加/编辑表单 */}
            {showAddForm && (
                <div className='form-overlay'>
                    <div className='form-container'>
                        <h3>{editingStock ? '编辑股票' : '添加股票'}</h3>
                        <form onSubmit={handleSubmit}>
                            <div className='form-group'>
                                <label>股票名称:</label>
                                <input
                                    type='text'
                                    value={formData.customName}
                                    onChange={(e) => setFormData({
                                        ...formData,
                                        customName: e.target.value
                                    })}
                                    placeholder='请输入股票名称'
                                    required
                                />
                            </div>
                            <div className='form-group'>
                                <label>股票代码:</label>
                                <input
                                    type='text'
                                    value={formData.customCode}
                                    onChange={(e) => setFormData({
                                        ...formData,
                                        customCode: e.target.value
                                    })}
                                    placeholder='例如: sh600000, sz000001, hk00700'
                                    required
                                />
                                <small className='form-hint'>
                                    支持格式: sh开头(上海)、sz开头(深圳)、hk开头(香港)
                                </small>
                            </div>
                            <div className='form-actions'>
                                <button 
                                    type='submit' 
                                    className='submit-button'
                                    disabled={loading}
                                >
                                    {loading ? '处理中...' : (editingStock ? '更新' : '添加')}
                                </button>
                                <button 
                                    type='button' 
                                    className='cancel-button'
                                    onClick={resetForm}
                                    disabled={loading}
                                >
                                    取消
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}

            {/* 股票列表 */}
            <div className='table-section'>
                <table className='custom-table'>
                    <thead>
                        <tr>
                            <th>名称</th>
                            <th>代码</th>
                            <th>价格</th>
                            <th>涨跌幅</th>
                            <th>涨跌额</th>
                            <th>成交量</th>
                            <th>成交额</th>
                            <th>公式计算结果</th>
                            <th>操作</th>
                        </tr>
                    </thead>
                    <tbody>
                        {loading ? (
                            <tr>
                                <td colSpan={9} className='loading-cell'>
                                    <div className='loading-content'>
                                        <span className='loading-spinner'></span>
                                        加载中...
                                    </div>
                                </td>
                            </tr>
                        ) : (
                            <>
                                {stocks.map(stock => (
                                    <tr key={stock.id}>
                                        <td className='stock-name'>{stock.customName}</td>
                                        <td className='stock-code'>{stock.customCode}</td>
                                        <td className='price'>{formatPrice(stock.price)}</td>
                                        <td>{formatChangeRate(stock.changeRate)}</td>
                                        <td className={stock.changeValue && stock.changeValue >= 0 ? 'positive' : 'negative'}>
                                            {stock.changeValue !== undefined ? stock.changeValue.toFixed(2) : '--'}
                                        </td>
                                        <td>{formatVolume(stock.volume)}</td>
                                        <td>{formatAmount(stock.amount)}</td>
                                        <td>{stock.formulaResult ? stock.formulaResult.toFixed(2) : '--'}</td>
                                        <td className='actions'>
                                            <button 
                                                className='edit-btn'
                                                onClick={() => handleEdit(stock)}
                                                disabled={loading}
                                            >
                                                编辑
                                            </button>
                                            <button 
                                                className='delete-btn'
                                                onClick={() => handleDelete(stock.id!)}
                                                disabled={loading}
                                            >
                                                删除
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                                {stocks.length === 0 && (
                                    <tr>
                                        <td colSpan={9} className='no-data-cell'>
                                            <div className='no-data'>
                                                <p>暂无自定义股票</p>
                                                <p>点击"添加股票"按钮开始添加</p>
                                            </div>
                                        </td>
                                    </tr>
                                )}
                            </>
                        )}
                    </tbody>
                </table>
            </div>
        </div>
    )
}

export default Custom
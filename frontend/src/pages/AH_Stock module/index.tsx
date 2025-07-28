import './AH_Stock.css'

function AH_Stock() {
    return (
        <div className='ah-stock-page'>
            <h2>AH_Stock</h2>

            {/* Search section */}
            <div className='search-section'>
                <input type='text' placeholder='Search' className='search-input'/>
                <button className='search-button'>Search ...</button>
            </div>

            {/* Table section */} 
            <div className='table-section'>
                <table>
                    <thead>
                        <tr>
                            <th>股票名称</th>
                            <th>A股股票代码</th>
                            <th>A股股票价格</th>
                            <th>更新时间</th>
                            <th>港股股票代码</th>
                            <th>港股股票价格</th>
                            <th>更新时间</th>
                        </tr>
                    </thead>
                </table>
            </div>
        </div>
    )
}

export default AH_Stock
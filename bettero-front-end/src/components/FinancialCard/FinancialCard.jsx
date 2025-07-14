function FinancialCard({ financialData }) {
  return (
    <>
      <h2 className="financial-info-title">This month's financial information</h2>
      <div className="financial-cards-wrapper">
        <div className="financial-card">
          <h4>Total balance:</h4>
          <div>${financialData.totalBalance}</div>
        </div>
        <div className="financial-card">
          <h4>Amount due:</h4>
          <div>${financialData.totalAmountDue}</div>
        </div>
        <div className="financial-card">
          <h4>Total income:</h4>
          <div>${financialData.totalIncome}</div>
        </div>
        <div className="financial-card">
          <h4>Total expense:</h4>
          <div>${financialData.totalExpense}</div>
        </div>
      </div>
    </>
  );
}

export default FinancialCard;
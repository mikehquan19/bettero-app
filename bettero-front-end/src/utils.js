/** 
 * Reformat the date for value of the date field in the form  
 */
export function reformatDate(argDate, fromDelimiter, toDelimiter) {
  if (argDate) {
    const dateArray = argDate.split(fromDelimiter);
    const month = dateArray[0];
    const day = dateArray[1];
    const year = dateArray[2];

    return (month + toDelimiter + day + toDelimiter + year);
  } else {
    return null;
  }
}

/**
 * Return the list of latest intervals of the predefined interval type
 */
export function latestIntervals(intervalContent, intervalType) {
  const latestIntervalList = [];
  const latestIntervalData = intervalContent[intervalType.toLowerCase()];

  latestIntervalData.forEach((interval) => {
    latestIntervalList.push({
      "first_date": interval.first_date,
      "last_date": interval.last_date
    });
  });

  return latestIntervalList;
}

/** 
 * Return the list of expense of latest intervals of predefined interval type 
 */
export function latestIntervalExpense(intervalContent, intervalType) {
  const intervalExpenseObj = {};
  const intervalExpenseData = intervalContent[intervalType.toLowerCase()];

  intervalExpenseData.forEach((expense) => {
    const firstDate = expense.first_date; 
    const totalExpense = expense.total_expense; 

    // add the total expense of the interval to the object 
    intervalExpenseObj[firstDate] = totalExpense;
  });

  return intervalExpenseObj;
}


/**
 * Return the object mapping the interval to change and composition
 */
export function latestIntervalChart(intervalContent, intervalType, firstDate) {
  const intervalObj = {};
  const intervalChartData = intervalContent[intervalType.toLowerCase()];

  intervalChartData.forEach((data) => {
    if (data.first_date === firstDate) {
      intervalObj = {
        "change_percentage": data.expense_change, 
        "composition_percentage": data.expense_composition,
        "daily_expense": data.daily_expense
      };
      
      return intervalObj;
    }
  });

  return null; 
}


import { Interval } from "@interface";

/**
 * Capitalize the word 
 * @param {string} word 
 * @returns {string}
 */
export function capitalize(word: string): string {
  return word.charAt(0).toUpperCase() + word.slice(1);
}

/**
 * Reformat the date for value of the date field in the form
 * @param {string} argDate 
 * @param {string} fromDelimiter 
 * @param {string} toDelimiter 
 * @returns {string}
 */
export function reformatDate(
  argDate: string, fromDelimiter: string, toDelimiter: string
): string {
  const dateArray = argDate.split(fromDelimiter);
  const month = dateArray[0];
  const day = dateArray[1];
  const year = dateArray[2];

  return (month + toDelimiter + day + toDelimiter + year);
}

/**
 * Return the list of latest intervals of the predefined interval type
 */
export function latestIntervals(intervalContent: any, intervalType: string): Interval[] {
  const latestIntervalList: Interval[] = [];
  const latestIntervalData = intervalContent[intervalType];

  latestIntervalData.forEach((interval: any) => {
    latestIntervalList.push({
      firstDate: interval.firstDate,
      lastDate: interval.lastDate
    });
  });

  return latestIntervalList;
}

/** 
 * Return the list of expense of latest intervals of predefined interval type 
 */
export function latestIntervalExpense(intervalContent: any, intervalType: string): Record<string, number> {
  const intervalExpenseObj: Record<string, number> = {};
  const intervalExpenseData = intervalContent[intervalType];

  intervalExpenseData.forEach((expense: any) => {
    const firstDate: string = expense.firstDate; 
    const totalExpense: number = expense.totalExpense; 

    // add the total expense of the interval to the object 
    intervalExpenseObj[firstDate] = totalExpense;
  });

  return intervalExpenseObj;
}


/**
 * Return the object mapping the interval to change and composition
 */
export function latestIntervalChart(
  intervalContent: any, intervalType: string, firstDate: string
) {
  const intervalChartData = intervalContent[intervalType];

  for (const data of intervalChartData) {
    if (data.firstDate === firstDate) {
      return {
        changePercentage: data.expenseChange, 
        compositionPercentage: data.expenseComposition,
        dailyExpense: data.dailyExpense
      };
    }
  }

  return null; 
}


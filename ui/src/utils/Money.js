import numeral from "numeral";
/**
 *
 * @param {float} amount amount to format
 * @returns formatted money with currency sign
 */
export const MoneyDirective = (amount) => {
  return numeral(amount).format("$ 0,0[.]00");
};

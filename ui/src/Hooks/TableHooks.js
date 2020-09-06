import { useState, useEffect } from "react";

export const useTableFilters = () => {
  const [state, setState] = useState({});

  // filter table was changed when the state (table properties) was changed
  useEffect(() => {
    const searchParams = new window.URLSearchParams(window.location.search);
    Object.keys(state).map((key) => {
      searchParams.set(key, state[key]);
    });

    // We don't want keep table action in browser back/forward history
    window.history.replaceState(
      null,
      null,
      decodeURIComponent(`?${searchParams.toString()}`)
    );
  }, [state]);

  const handleChange = (filters = []) => {
    const newFilters = {};
    filters.forEach((filter) => (newFilters[filter.key] = filter.value));
    Object.assign(state, newFilters);
    setState({ ...state });
  };
  return [handleChange];
};

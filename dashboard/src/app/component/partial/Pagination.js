import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withStyles } from "@material-ui/core/styles";

import TablePagination from "@material-ui/core/TablePagination";

const StyledTablePagination = withStyles(theme => ({
  root: {
    position: "sticky",
    bottom: 0,
    backgroundColor: theme.palette.background.paper,
    borderTopColor: theme.palette.divider,
    borderTopWidth: 1,
    borderTopStyle: "solid",
    marginTop: -1,
  },
}))(TablePagination);

class Pagination extends React.PureComponent {
  static propTypes = {
    pageInfo: PropTypes.object,
    limit: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    offset: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    onChangeQuery: PropTypes.func,
  };

  static defaultProps = {
    pageInfo: undefined,
    limit: undefined,
    offset: undefined,
    onChangeQuery: undefined,
  };

  static fragments = {
    pageInfo: gql`
      fragment Pagination_pageInfo on OffsetPageInfo {
        totalCount
      }
    `,
  };

  _handleChangePage = (event, page) => {
    if (this.props.onChangeQuery) {
      this.props.onChangeQuery({ offset: page * this.props.limit });
    }
  };

  _handleChangeLimit = event => {
    if (this.props.onChangeQuery) {
      this.props.onChangeQuery({ limit: event.target.value });
    }
  };

  render() {
    const { limit: rawLimit, offset: rawOffset, pageInfo } = this.props;

    const limit = parseInt(rawLimit, 10);
    const offset = parseInt(rawOffset, 10);

    if (Number.isNaN(limit)) {
      throw new TypeError(`Expected numeric limit. Received ${rawLimit}`);
    }

    if (Number.isNaN(offset)) {
      throw new TypeError(`Expected numeric offset. Received ${rawOffset}`);
    }

    // Fall back to a placeholder total count value (equal to page size) while
    // pageInfo is undefined during load.
    const count = pageInfo ? pageInfo.totalCount : limit;

    // Given that offset isn't strictly a multiple of limit, the current page
    // index must be rounded down. In the case that we have a small offset that
    // would otherwise round to zero, return 1 to enable the previous page
    // button to reset offset to 0.
    const page = offset > 0 && offset < limit ? 1 : Math.floor(offset / limit);

    return (
      <StyledTablePagination
        component="div"
        count={count}
        rowsPerPage={limit}
        page={page}
        onChangePage={this._handleChangePage}
        onChangeRowsPerPage={this._handleChangeLimit}
        rowsPerPageOptions={[5, 10, 25, 50, 100, 200]}
      />
    );
  }
}

export default Pagination;

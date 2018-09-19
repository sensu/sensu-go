import React from "react";
import PropTypes from "prop-types";

import ListToolbar from "/components/partials/ListToolbar";
import ResetMenuItem from "/components/partials/ToolbarMenuItems/Reset";
import SearchBox from "/components/SearchBox";
import ToolbarMenu from "/components/partials/ToolbarMenu";

class ChecksListToolbar extends React.PureComponent {
  static propTypes = {
    query: PropTypes.string,
    onChangeQuery: PropTypes.func.isRequired,
    onClickReset: PropTypes.func.isRequired,
  };

  static defaultProps = {
    query: "",
  };

  render() {
    const { onChangeQuery, onClickReset, query } = this.props;

    return (
      <ListToolbar
        search={
          <SearchBox
            placeholder="Filter checks…"
            initialValue={query}
            onSearch={onChangeQuery}
          />
        }
        toolbarItems={({ collapsed }) => (
          <ToolbarMenu>
            <ToolbarMenu.Item
              id="reset"
              visible={collapsed ? "never" : "if-room"}
            >
              <ResetMenuItem onClick={onClickReset} />
            </ToolbarMenu.Item>
          </ToolbarMenu>
        )}
      />
    );
  }
}

export default ChecksListToolbar;

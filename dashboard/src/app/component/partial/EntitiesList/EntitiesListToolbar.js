import React from "react";
import PropTypes from "prop-types";

import ListToolbar from "/app/component/partial/ListToolbar";
import ResetMenuItem from "/app/component/partial/ToolbarMenuItems/Reset";
import SearchBox from "/lib/component/base/SearchBox";
import ToolbarMenu from "/app/component/partial/ToolbarMenu";

class EntitiesListToolbar extends React.PureComponent {
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
            placeholder="Filter entities…"
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

export default EntitiesListToolbar;

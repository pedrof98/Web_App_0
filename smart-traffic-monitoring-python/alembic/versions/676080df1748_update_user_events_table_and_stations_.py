"""Update user_events table and stations table

Revision ID: 676080df1748
Revises: 743cf310e60e
Create Date: 2025-01-03 08:18:21.396233

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '676080df1748'
down_revision: Union[str, None] = '743cf310e60e'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # ### commands auto generated by Alembic - please adjust! ###
    op.add_column('user_events', sa.Column('station_id', sa.Integer(), nullable=True))
    op.create_foreign_key(None, 'user_events', 'stations', ['station_id'], ['id'])
    # ### end Alembic commands ###


def downgrade() -> None:
    # ### commands auto generated by Alembic - please adjust! ###
    op.drop_constraint(None, 'user_events', type_='foreignkey')
    op.drop_column('user_events', 'station_id')
    # ### end Alembic commands ###

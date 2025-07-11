<?php

use yii\db\Migration;

/**
 * Class m241205_105659_add_firest_migration
 */
class m241205_105659_add_firest_migration extends Migration
{
    /**
     * {@inheritdoc}
     */
    public function safeUp()
    {
        // $path = __DIR__.'/../data/library.db';
        // $db = new SQLite3($path);
        
        $this->createTable('params', [
            'id' => $this->primaryKey(),
            'param' => $this->string(255)->notNull(),
            'value' => $this->string(255)->notNull(),
        ]);
    }

    /**
     * {@inheritdoc}
     */
    public function safeDown()
    {
        echo "m241205_105659_add_firest_migration cannot be reverted.\n";

        return false;
    }

    /*
    // Use up()/down() to run migration code without a transaction.
    public function up()
    {

    }

    public function down()
    {
        echo "m241205_105659_add_firest_migration cannot be reverted.\n";

        return false;
    }
    */
}
